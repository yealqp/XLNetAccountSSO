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
	authService         *service.AuthService
	adminService        *service.AdminService
	verificationService *service.VerificationService
	cfg                 config.Config
}

func NewAuthHandler(authService *service.AuthService, adminService *service.AdminService, verificationService *service.VerificationService, cfg config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, adminService: adminService, verificationService: verificationService, cfg: cfg}
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
		if errors.Is(err, service.ErrConflict) {
			return writeError(c, fiber.StatusConflict, "系统尚未初始化，请先创建管理员账号")
		}
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

func (handler *AuthHandler) Register(c *fiber.Ctx) error {
	var input service.RegisterInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid register payload")
	}
	user, err := handler.authService.Register(context.Background(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConflict):
			return writeError(c, fiber.StatusConflict, "邮箱已被注册，或系统尚未初始化")
		case errors.Is(err, service.ErrForbidden):
			return writeError(c, fiber.StatusForbidden, "当前未开放注册")
		case errors.Is(err, service.ErrInvalidInput):
			return writeError(c, fiber.StatusBadRequest, "请输入有效邮箱、密码和验证码")
		case errors.Is(err, service.ErrInvalidGrant):
			return writeError(c, fiber.StatusBadRequest, "验证码错误或已过期")
		default:
			return writeError(c, fiber.StatusInternalServerError, "register failed")
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"registered": true,
		"user":       publicUser(user),
	})
}

func (handler *AuthHandler) SendRegisterCode(c *fiber.Ctx) error {
	var input struct {
		Email        string `json:"email"`
		CaptchaToken string `json:"captcha_token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid register code payload")
	}
	settings, err := handler.adminService.PlatformSettings(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load registration settings failed")
	}
	if err := handler.verificationService.SendRegistrationCode(context.Background(), settings, input.Email, input.CaptchaToken); err != nil {
		switch {
		case errors.Is(err, service.ErrConflict):
			return writeError(c, fiber.StatusConflict, "邮箱已被注册")
		case errors.Is(err, service.ErrForbidden):
			return writeError(c, fiber.StatusForbidden, "当前未开放注册")
		case errors.Is(err, service.ErrInvalidInput):
			return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidInput))
		default:
			return writeError(c, fiber.StatusInternalServerError, "send register code failed")
		}
	}
	return c.JSON(fiber.Map{"sent": true})
}

func (handler *AuthHandler) SetupStatus(c *fiber.Ctx) error {
	initialized, err := handler.authService.IsInitialized(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load setup status failed")
	}
	return c.JSON(fiber.Map{"initialized": initialized})
}

func (handler *AuthHandler) Initialize(c *fiber.Ctx) error {
	var input service.InitializeAdminInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid setup payload")
	}
	user, err := handler.authService.InitializeFirstAdmin(context.Background(), input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConflict):
			return writeError(c, fiber.StatusConflict, "系统已经初始化")
		case errors.Is(err, service.ErrInvalidInput):
			return writeError(c, fiber.StatusBadRequest, "用户名和密码不能为空")
		default:
			return writeError(c, fiber.StatusInternalServerError, "initialize admin failed")
		}
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"initialized": true,
		"user":        publicUser(user),
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
