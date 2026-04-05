package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/xianlin-network/sso-platform/server/internal/config"
	"github.com/xianlin-network/sso-platform/server/internal/http/middleware"
	"github.com/xianlin-network/sso-platform/server/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, cfg: cfg}
}

func (handler *AuthHandler) Login(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid login payload")
	}
	user, rawSessionToken, session, err := handler.authService.Login(context.Background(), input.Username, input.Password, service.SessionMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			return writeError(c, fiber.StatusUnauthorized, "username or password is incorrect")
		}
		return writeError(c, fiber.StatusInternalServerError, "login failed")
	}

	c.Cookie(&fiber.Cookie{
		Name:     handler.cfg.CookieName,
		Value:    rawSessionToken,
		HTTPOnly: true,
		Secure:   handler.cfg.CookieSecure,
		SameSite: "lax",
		Path:     "/",
		Expires:  session.ExpiresAt,
	})

	return c.JSON(fiber.Map{
		"authenticated": true,
		"user":          publicUser(user),
	})
}

func (handler *AuthHandler) Logout(c *fiber.Ctx) error {
	_ = handler.authService.Logout(context.Background(), c.Cookies(handler.cfg.CookieName))
	c.Cookie(&fiber.Cookie{
		Name:     handler.cfg.CookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   handler.cfg.CookieSecure,
		SameSite: "lax",
		Path:     "/",
		Expires:  time.Unix(0, 0),
	})
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *AuthHandler) Session(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return c.JSON(fiber.Map{"authenticated": false})
	}
	return c.JSON(fiber.Map{
		"authenticated": true,
		"user":          publicUser(authContext.User),
	})
}
