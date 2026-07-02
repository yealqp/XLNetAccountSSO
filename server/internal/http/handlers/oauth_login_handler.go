package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/http/middleware"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/pkg/security"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

type OAuthLoginHandler struct {
	oauthLoginService *service.OAuthLoginService
	cfg               config.Config
}

func NewOAuthLoginHandler(oauthLoginService *service.OAuthLoginService, cfg config.Config) *OAuthLoginHandler {
	return &OAuthLoginHandler{oauthLoginService: oauthLoginService, cfg: cfg}
}

const stateTTL = 10 * time.Minute

type stateEntry struct {
	expiresAt time.Time
	binding   bool   // true = linking to existing user, false = login flow
	token     string // session token for binding auth
}

var oauthStates = struct {
	m map[string]stateEntry
}{
	m: make(map[string]stateEntry),
}

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			now := time.Now()
			for k, v := range oauthStates.m {
				if now.After(v.expiresAt) {
					delete(oauthStates.m, k)
				}
			}
		}
	}()
}

func generateState(binding bool, token ...string) string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := hex.EncodeToString(b)
	tok := ""
	if len(token) > 0 {
		tok = token[0]
	}
	oauthStates.m[state] = stateEntry{expiresAt: time.Now().Add(stateTTL), binding: binding, token: tok}
	return state
}

func consumeState(state string) (stateEntry, bool) {
	entry, ok := oauthStates.m[state]
	if !ok {
		return stateEntry{}, false
	}
	delete(oauthStates.m, state)
	return entry, true
}

func (h *OAuthLoginHandler) OAuthLogin(c *fiber.Ctx) error {
	provider := service.OAuthProvider(strings.TrimSpace(c.Params("provider")))
	state := generateState(false)

	authURL, err := h.oauthLoginService.AuthorizeURL(provider, state)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Redirect(authURL, fiber.StatusFound)
}

func (h *OAuthLoginHandler) OAuthBind(c *fiber.Ctx) error {
	authCtx, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}

	provider := service.OAuthProvider(strings.TrimSpace(c.Params("provider")))
	state := generateState(true, authCtx.Token)

	authURL, err := h.oauthLoginService.AuthorizeURL(provider, state)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Redirect(authURL, fiber.StatusFound)
}

func (h *OAuthLoginHandler) OAuthCallback(c *fiber.Ctx) error {
	provider := service.OAuthProvider(strings.TrimSpace(c.Params("provider")))
	code := strings.TrimSpace(c.Query("code"))
	stateStr := strings.TrimSpace(c.Query("state"))
	errorParam := strings.TrimSpace(c.Query("error"))

	if errorParam != "" {
		slog.Warn("oauth callback: provider returned error",
			slog.String("provider", string(provider)),
			slog.String("error", errorParam),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error="+errorParam, fiber.StatusFound)
	}

	state, ok := consumeState(stateStr)
	if code == "" || !ok || time.Now().After(state.expiresAt) {
		slog.Warn("oauth callback: invalid code or state",
			slog.String("provider", string(provider)),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=invalid_request", fiber.StatusFound)
	}

	info, err := h.oauthLoginService.ExchangeCode(context.Background(), provider, code)
	if err != nil {
		slog.Warn("oauth exchange failed",
			slog.String("provider", string(provider)),
			slog.String("error", err.Error()),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=exchange_failed", fiber.StatusFound)
	}

	if state.binding {
		return h.handleBindCallback(c, provider, info, state.token)
	}
	return h.handleLoginCallback(c, provider, info)
}

func (h *OAuthLoginHandler) handleLoginCallback(c *fiber.Ctx, provider service.OAuthProvider, info *service.OAuthUserInfo) error {
	user, rawToken, _, err := h.oauthLoginService.FindOrCreateUser(context.Background(), provider, info)
	if err != nil {
		slog.Error("oauth login failed",
			slog.String("provider", string(provider)),
			slog.String("error", err.Error()),
		)
		if errors.Is(err, service.ErrNotFound) {
			return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=not_found", fiber.StatusFound)
		}
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=login_failed", fiber.StatusFound)
	}

	slog.Info("oauth login success",
		slog.String("provider", string(provider)),
		slog.String("username", user.Username),
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("ip", c.IP()),
	)

	target := h.cfg.WebBaseURL + "/auth/login?oauth_token=" + rawToken
	return c.Redirect(target, fiber.StatusFound)
}

func (h *OAuthLoginHandler) handleBindCallback(c *fiber.Ctx, provider service.OAuthProvider, info *service.OAuthUserInfo, bindToken string) error {
	bindUser, _, err := h.oauthLoginService.ResolveSession(context.Background(), bindToken)
	if err != nil || bindUser == nil {
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=auth_required", fiber.StatusFound)
	}

	// Check if already bound to another user
	existing, _ := h.oauthLoginService.FindLinkedAccount(context.Background(), provider, info.ID)
	if existing != nil {
		return c.Redirect(h.cfg.WebBaseURL+"/dashboard/settings?oauth_bind_error=already_bound", fiber.StatusFound)
	}

	// Check if already bound to this user
	existing, _ = h.oauthLoginService.FindLinkedAccountByUser(context.Background(), provider, bindUser.ID)
	if existing != nil {
		return c.Redirect(h.cfg.WebBaseURL+"/dashboard/settings?oauth_bind_error=already_bound_to_user", fiber.StatusFound)
	}

	link := &model.OAuthLinkedAccount{
		ID:             security.NewID(),
		UserID:         bindUser.ID,
		Provider:       string(provider),
		ProviderUserID: info.ID,
		Email:          info.Email,
		Name:           info.Name,
	}
	if err := h.oauthLoginService.LinkAccount(context.Background(), link); err != nil {
		slog.Error("oauth bind failed",
			slog.String("provider", string(provider)),
			slog.String("error", err.Error()),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/dashboard/settings?oauth_bind_error=link_failed", fiber.StatusFound)
	}

	slog.Info("oauth bind success",
		slog.String("provider", string(provider)),
		slog.Uint64("user_id", uint64(bindUser.ID)),
	)
	return c.Redirect(h.cfg.WebBaseURL+"/dashboard/settings?oauth_bind_success="+string(provider), fiber.StatusFound)
}

func (h *OAuthLoginHandler) ListBindings(c *fiber.Ctx) error {
	authCtx, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}

	accounts, err := h.oauthLoginService.ListAccounts(context.Background(), authCtx.User.ID)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load bindings failed")
	}

	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": accounts}, "success")
}

func (h *OAuthLoginHandler) UnlinkBinding(c *fiber.Ctx) error {
	authCtx, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}

	id := c.Params("id")
	if err := h.oauthLoginService.UnlinkAccount(context.Background(), id, authCtx.User.ID); err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}

	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}
