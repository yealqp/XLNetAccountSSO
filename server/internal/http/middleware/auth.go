package middleware

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/xianlin-network/sso-platform/server/internal/model"
	"github.com/xianlin-network/sso-platform/server/internal/service"
)

const authContextKey = "auth_context"

type AuthContext struct {
	User    *model.User
	Session *model.UserSession
}

func OptionalSession(authService *service.AuthService, cookieName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawSession := c.Cookies(cookieName)
		if rawSession == "" {
			return c.Next()
		}
		user, session, err := authService.ResolveSession(context.Background(), rawSession)
		if err != nil {
			return c.Next()
		}
		c.Locals(authContextKey, &AuthContext{User: user, Session: session})
		return c.Next()
	}
}

func RequireSession(authService *service.AuthService, cookieName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawSession := c.Cookies(cookieName)
		user, session, err := authService.ResolveSession(context.Background(), rawSession)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		c.Locals(authContextKey, &AuthContext{User: user, Session: session})
		return c.Next()
	}
}

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authContext, err := CurrentAuthContext(c)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		if authContext.User.Role != "admin" {
			return fiber.NewError(fiber.StatusForbidden, "admin access required")
		}
		return c.Next()
	}
}

func CurrentAuthContext(c *fiber.Ctx) (*AuthContext, error) {
	value := c.Locals(authContextKey)
	authContext, ok := value.(*AuthContext)
	if !ok || authContext == nil || authContext.User == nil {
		return nil, errors.New("auth context unavailable")
	}
	return authContext, nil
}
