package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

const authContextKey = "auth_context"

type AuthContext struct {
	User    *model.User
	Session *model.UserSession
	Token   string
}

func OptionalSession(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawSession := ResolveAuthToken(c)
		if rawSession == "" {
			return c.Next()
		}
		user, session, err := authService.ResolveSession(context.Background(), rawSession)
		if err != nil {
			return c.Next()
		}
		c.Locals(authContextKey, &AuthContext{User: user, Session: session, Token: rawSession})
		return c.Next()
	}
}

func RequireSession(authService *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawSession := ResolveAuthToken(c)
		user, session, err := authService.ResolveSession(context.Background(), rawSession)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}
		c.Locals(authContextKey, &AuthContext{User: user, Session: session, Token: rawSession})
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

func ResolveAuthToken(c *fiber.Ctx) string {
	authorization := strings.TrimSpace(c.Get(fiber.HeaderAuthorization))
	if authorization != "" && strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		return strings.TrimSpace(authorization[7:])
	}
	return ""
}
