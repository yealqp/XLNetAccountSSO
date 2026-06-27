package handlers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/http/middleware"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService         *service.AuthService
	passkeyService      *service.PasskeyService
	adminService        *service.AdminService
	verificationService *service.VerificationService
	cfg                 config.Config
}

func NewAuthHandler(authService *service.AuthService, passkeyService *service.PasskeyService, adminService *service.AdminService, verificationService *service.VerificationService, cfg config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, passkeyService: passkeyService, adminService: adminService, verificationService: verificationService, cfg: cfg}
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
			slog.Warn("login failed: invalid credentials",
				slog.String("username", input.Username),
				slog.String("ip", c.IP()),
			)
			return writeError(c, fiber.StatusUnauthorized, "username or password is incorrect")
		}
		return writeError(c, fiber.StatusInternalServerError, "login failed")
	}

	slog.Info("login success",
		slog.String("username", user.Username),
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("ip", c.IP()),
	)

	return writeSuccess(c, fiber.StatusOK, authSessionPayload(user, rawSessionToken, session), "success")
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
			return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidInput))
		case errors.Is(err, service.ErrInvalidGrant):
			return writeError(c, fiber.StatusBadRequest, "验证码错误或已过期")
		default:
			return writeError(c, fiber.StatusInternalServerError, "register failed")
		}
	}

	slog.Info("register success",
		slog.String("username", user.Username),
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("email", user.Email),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusCreated, fiber.Map{
		"registered": true,
		"user":       publicUser(user),
	}, "success")
}

func (handler *AuthHandler) UpdateProfile(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	var input service.UpdateProfileInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid profile payload")
	}
	user, err := handler.authService.UpdateProfile(context.Background(), authContext.User, input)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrConflict):
			return writeError(c, fiber.StatusConflict, "用户名已被占用")
		case errors.Is(err, service.ErrInvalidInput):
			return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidInput))
		case errors.Is(err, service.ErrUnauthorized):
			return writeError(c, fiber.StatusUnauthorized, "authentication required")
		case errors.Is(err, service.ErrInvalidGrant):
			return writeError(c, fiber.StatusBadRequest, "验证码错误或已过期")
		default:
			return writeError(c, fiber.StatusInternalServerError, "update profile failed")
		}
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"user": publicUser(user)}, "success")
}

func (handler *AuthHandler) SendProfilePasswordCode(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	if err := handler.verificationService.SendProfilePasswordCode(context.Background(), handler.cfg, authContext.User.Email); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInput):
			return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidInput))
		default:
			return writeError(c, fiber.StatusInternalServerError, "send profile password code failed")
		}
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"sent": true}, "success")
}

func (handler *AuthHandler) SendRegisterCode(c *fiber.Ctx) error {
	var input struct {
		Email        string `json:"email"`
		CaptchaToken string `json:"captcha_token"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid register code payload")
	}
	if err := handler.verificationService.SendRegistrationCode(context.Background(), handler.cfg, input.Email, input.CaptchaToken); err != nil {
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
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"sent": true}, "success")
}

func (handler *AuthHandler) SetupStatus(c *fiber.Ctx) error {
	initialized, err := handler.authService.IsInitialized(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load setup status failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"initialized": initialized}, "success")
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

	slog.Info("system initialized: first admin created",
		slog.String("username", user.Username),
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusCreated, fiber.Map{
		"initialized": true,
		"user":        publicUser(user),
	}, "success")
}

func (handler *AuthHandler) Logout(c *fiber.Ctx) error {
	_ = handler.authService.Logout(context.Background(), middleware.ResolveAuthToken(c))
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AuthHandler) Session(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeSuccess(c, fiber.StatusOK, fiber.Map{"authenticated": false}, "success")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{
		"authenticated": true,
		"user":          publicUser(authContext.User),
	}, "success")
}

func authSessionPayload(user *model.User, rawSessionToken string, session *model.UserSession) fiber.Map {
	return fiber.Map{
		"authenticated": true,
		"access_token":  rawSessionToken,
		"token_type":    "Bearer",
		"expires_in":    int(time.Until(session.ExpiresAt).Seconds()),
		"user":          publicUser(user),
	}
}
