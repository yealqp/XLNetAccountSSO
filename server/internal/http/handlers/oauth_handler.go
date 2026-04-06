package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/http/middleware"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
	cfg          config.Config
}

func NewOAuthHandler(oauthService *service.OAuthService, cfg config.Config) *OAuthHandler {
	return &OAuthHandler{oauthService: oauthService, cfg: cfg}
}

func (handler *OAuthHandler) AuthorizeEntry(c *fiber.Ctx) error {
	query := string(c.Request().URI().QueryString())
	redirectURL := handler.cfg.WebBaseURL + "/auth/authorize"
	if query != "" {
		redirectURL += "?" + query
	}
	return c.Redirect(redirectURL, fiber.StatusTemporaryRedirect)
}

func (handler *OAuthHandler) OpenIDConfiguration(c *fiber.Ctx) error {
	return c.JSON(handler.oauthService.DiscoveryMetadata())
}

func (handler *OAuthHandler) JWKS(c *fiber.Ctx) error {
	return c.JSON(handler.oauthService.JSONWebKeySet())
}

func (handler *OAuthHandler) Preview(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	preview, err := handler.oauthService.PreviewAuthorization(context.Background(), authContext.User, service.AuthorizationRequest{
		ResponseType:        c.Query("response_type"),
		ClientID:            c.Query("client_id"),
		RedirectURI:         c.Query("redirect_uri"),
		Scope:               c.Query("scope"),
		State:               c.Query("state"),
		Nonce:               c.Query("nonce"),
		CodeChallenge:       c.Query("code_challenge"),
		CodeChallengeMethod: c.Query("code_challenge_method"),
	})
	if err != nil {
		return handleServiceError(c, err, "preview authorization failed")
	}
	return writeSuccess(c, fiber.StatusOK, preview, "success")
}

func (handler *OAuthHandler) Decide(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	var input service.AuthorizationDecision
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid authorization payload")
	}
	redirectTo, err := handler.oauthService.Authorize(context.Background(), authContext.User, input)
	if err != nil {
		return handleServiceError(c, err, "issue authorization code failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"redirect_to": redirectTo}, "success")
}

func (handler *OAuthHandler) Token(c *fiber.Ctx) error {
	clientID, clientSecret := parseClientCredentials(c)
	response, err := handler.oauthService.ExchangeToken(context.Background(), service.TokenRequest{
		GrantType:    c.FormValue("grant_type"),
		Code:         c.FormValue("code"),
		RedirectURI:  c.FormValue("redirect_uri"),
		ClientID:     fallbackValue(c.FormValue("client_id"), clientID),
		ClientSecret: fallbackValue(c.FormValue("client_secret"), clientSecret),
		CodeVerifier: c.FormValue("code_verifier"),
		RefreshToken: c.FormValue("refresh_token"),
	})
	if err != nil {
		return handleOAuthError(c, err, "exchange token failed")
	}
	c.Set(fiber.HeaderCacheControl, "no-store")
	c.Set("Pragma", "no-cache")
	return c.JSON(response)
}

func (handler *OAuthHandler) UserInfo(c *fiber.Ctx) error {
	token := bearerToken(c)
	info, err := handler.oauthService.UserInfo(context.Background(), token)
	if err != nil {
		return handleOAuthError(c, err, "load userinfo failed")
	}
	return c.JSON(info)
}

func (handler *OAuthHandler) Revoke(c *fiber.Ctx) error {
	clientID, _ := parseClientCredentials(c)
	err := handler.oauthService.RevokeToken(context.Background(), c.FormValue("token"), fallbackValue(c.FormValue("client_id"), clientID))
	if err != nil {
		return handleOAuthError(c, err, "revoke token failed")
	}
	return c.SendStatus(fiber.StatusOK)
}

func (handler *OAuthHandler) Introspect(c *fiber.Ctx) error {
	result, err := handler.oauthService.Introspect(context.Background(), c.FormValue("token"))
	if err != nil {
		return handleOAuthError(c, err, "introspect token failed")
	}
	return c.JSON(result)
}

func parseClientCredentials(c *fiber.Ctx) (string, string) {
	username, password, ok := service.DecodeBasicAuth(c.Get(fiber.HeaderAuthorization))
	if !ok {
		return "", ""
	}
	return username, password
}

func fallbackValue(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

func bearerToken(c *fiber.Ctx) string {
	return strings.TrimSpace(strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer "))
}

func handleOAuthError(c *fiber.Ctx, err error, fallback string) error {
	switch {
	case errors.Is(err, service.ErrInvalidClient):
		c.Set(fiber.HeaderWWWAuthenticate, `Basic realm="oauth"`)
		return writeOAuthError(c, fiber.StatusUnauthorized, "invalid_client", err.Error())
	case errors.Is(err, service.ErrInvalidGrant):
		return writeOAuthError(c, fiber.StatusBadRequest, "invalid_grant", err.Error())
	case errors.Is(err, service.ErrInvalidInput):
		return writeOAuthError(c, fiber.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, service.ErrInvalidToken):
		c.Set(fiber.HeaderWWWAuthenticate, `Bearer error="invalid_token", error_description="invalid token"`)
		return writeOAuthError(c, fiber.StatusUnauthorized, "invalid_token", err.Error())
	case errors.Is(err, service.ErrAccessDenied):
		c.Set(fiber.HeaderWWWAuthenticate, `Bearer error="insufficient_scope", error_description="openid scope required"`)
		return writeOAuthError(c, fiber.StatusForbidden, "insufficient_scope", err.Error())
	case errors.Is(err, service.ErrUnauthorized):
		return writeOAuthError(c, fiber.StatusUnauthorized, "invalid_request", err.Error())
	default:
		return writeOAuthError(c, fiber.StatusInternalServerError, "server_error", fallback)
	}
}

func writeOAuthError(c *fiber.Ctx, status int, code string, description string) error {
	response := fiber.Map{"error": code}
	if strings.TrimSpace(description) != "" {
		response["error_description"] = description
		response["message"] = description
	}
	return c.Status(status).JSON(response)
}
