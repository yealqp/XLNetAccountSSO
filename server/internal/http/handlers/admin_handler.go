package handlers

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/XianLinNet/XLNetAccount/internal/http/middleware"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	adminService *service.AdminService
	tokenService *service.TokenService
}

func NewAdminHandler(adminService *service.AdminService, tokenService *service.TokenService) *AdminHandler {
	return &AdminHandler{adminService: adminService, tokenService: tokenService}
}

func (handler *AdminHandler) Overview(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	overview, err := handler.adminService.OverviewForUser(context.Background(), authContext.User)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load overview failed")
	}
	return writeSuccess(c, fiber.StatusOK, overview, "success")
}

func (handler *AdminHandler) AdminOverview(c *fiber.Ctx) error {
	overview, err := handler.adminService.Overview(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load overview failed")
	}
	return writeSuccess(c, fiber.StatusOK, overview, "success")
}

func (handler *AdminHandler) PublicSettings(c *fiber.Ctx) error {
	settings, err := handler.adminService.PublicSettings(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load public settings failed")
	}
	return writeSuccess(c, fiber.StatusOK, settings, "success")
}

func (handler *AdminHandler) Me(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"user": publicUser(authContext.User)}, "success")
}

func (handler *AdminHandler) PlatformSettings(c *fiber.Ctx) error {
	settings, err := handler.adminService.PlatformSettings(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load platform settings failed")
	}
	return writeSuccess(c, fiber.StatusOK, settings, "success")
}

func (handler *AdminHandler) UpdatePlatformSettings(c *fiber.Ctx) error {
	var input service.PlatformSettings
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid settings payload")
	}
	settings, err := handler.adminService.UpdatePlatformSettings(context.Background(), input)
	if err != nil {
		return handleServiceError(c, err, "update platform settings failed")
	}
	slog.Info("platform settings updated",
		slog.String("platform_name", input.PlatformName),
		slog.Bool("allow_registration", input.AllowRegistration),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusOK, settings, "success")
}

func (handler *AdminHandler) SendTestEmail(c *fiber.Ctx) error {
	var input service.TestEmailInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid test email payload")
	}
	if err := handler.adminService.SendTestEmail(context.Background(), input); err != nil {
		return handleServiceError(c, err, "send test email failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"sent": true}, "success")
}

func (handler *AdminHandler) ListClients(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	clients, err := handler.adminService.ListUserClients(context.Background(), authContext.User)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load clients failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": clients}, "success")
}

func (handler *AdminHandler) ListManagedClients(c *fiber.Ctx) error {
	clients, err := handler.adminService.ListClients(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load clients failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": clients}, "success")
}

func (handler *AdminHandler) CreateClient(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	var input service.CreateClientInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid client payload")
	}
	client, err := handler.adminService.CreateClient(context.Background(), authContext.User, input)
	if err != nil {
		return handleServiceError(c, err, "create client failed")
	}
	slog.Info("client created",
		slog.String("client_name", input.Name),
		slog.String("client_id", input.ClientID),
		slog.String("by", authContext.User.Username),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusCreated, client, "success")
}

func (handler *AdminHandler) UploadClientIcon(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "请选择要上传的图标文件")
	}
	result, err := handler.adminService.UploadClientIcon(context.Background(), fileHeader)
	if err != nil {
		return handleServiceError(c, err, "upload client icon failed")
	}
	return writeSuccess(c, fiber.StatusCreated, result, "success")
}

func (handler *AdminHandler) UploadWebIcon(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, "请选择要上传的图标文件")
	}
	result, err := handler.adminService.UploadWebIcon(context.Background(), fileHeader)
	if err != nil {
		return handleServiceError(c, err, "upload web icon failed")
	}
	return writeSuccess(c, fiber.StatusCreated, result, "success")
}

func (handler *AdminHandler) UpdateClient(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	var input service.UpdateClientInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid client payload")
	}
	client, err := handler.adminService.UpdateUserClient(context.Background(), authContext.User, c.Params("id"), input)
	if err != nil {
		return handleServiceError(c, err, "update client failed")
	}
	return writeSuccess(c, fiber.StatusOK, client, "success")
}

func (handler *AdminHandler) UpdateManagedClient(c *fiber.Ctx) error {
	var input service.UpdateClientInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid client payload")
	}
	client, err := handler.adminService.UpdateClient(context.Background(), c.Params("id"), input)
	if err != nil {
		return handleServiceError(c, err, "update client failed")
	}
	return writeSuccess(c, fiber.StatusOK, client, "success")
}

func (handler *AdminHandler) DeleteClient(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	err = handler.adminService.DeleteUserClient(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "delete client failed")
	}
	slog.Info("client deleted",
		slog.String("client_id", c.Params("id")),
		slog.String("by", authContext.User.Username),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AdminHandler) DeleteManagedClient(c *fiber.Ctx) error {
	err := handler.adminService.DeleteClient(context.Background(), c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "delete client failed")
	}
	slog.Info("admin deleted client",
		slog.String("client_id", c.Params("id")),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := handler.adminService.ListUsers(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load users failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": users}, "success")
}

func (handler *AdminHandler) CreateUser(c *fiber.Ctx) error {
	var input service.CreateUserInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid user payload")
	}
	user, err := handler.adminService.CreateUser(context.Background(), input)
	if err != nil {
		return handleServiceError(c, err, "create user failed")
	}
	slog.Info("admin created user",
		slog.String("username", input.Username),
		slog.String("role", input.Role),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusCreated, user, "success")
}

func (handler *AdminHandler) UpdateUser(c *fiber.Ctx) error {
	var input service.UpdateUserInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid user payload")
	}
	user, err := handler.adminService.UpdateUser(context.Background(), c.Params("id"), input)
	if err != nil {
		return handleServiceError(c, err, "update user failed")
	}
	return writeSuccess(c, fiber.StatusOK, user, "success")
}

func (handler *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.adminService.DeleteUser(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "delete user failed")
	}
	slog.Info("admin deleted user",
		slog.String("user_id", c.Params("id")),
		slog.String("by", authContext.User.Username),
		slog.String("ip", c.IP()),
	)
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AdminHandler) ListTokens(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	tokens, err := handler.tokenService.ListTokens(context.Background(), authContext.User)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load tokens failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": tokens}, "success")
}

func (handler *AdminHandler) ListManagedTokens(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	tokens, err := handler.tokenService.ListAllTokens(context.Background(), authContext.User)
	if err != nil {
		return handleServiceError(c, err, "load tokens failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": tokens}, "success")
}

func (handler *AdminHandler) RevokeAccessToken(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.tokenService.RevokeAccessToken(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "revoke access token failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AdminHandler) RevokeRefreshToken(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.tokenService.RevokeRefreshToken(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "revoke refresh token failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AdminHandler) RevokeClientTokens(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.tokenService.RevokeClientTokens(context.Background(), authContext.User, c.Params("clientId"))
	if err != nil {
		return handleServiceError(c, err, "revoke client tokens failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func handleServiceError(c *fiber.Ctx, err error, fallback string) error {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidInput))
	case errors.Is(err, service.ErrConflict):
		return writeError(c, fiber.StatusConflict, cleanServiceError(err, service.ErrConflict))
	case errors.Is(err, service.ErrNotFound):
		return writeError(c, fiber.StatusNotFound, cleanServiceError(err, service.ErrNotFound))
	case errors.Is(err, service.ErrForbidden):
		return writeError(c, fiber.StatusForbidden, cleanServiceError(err, service.ErrForbidden))
	case errors.Is(err, service.ErrUnauthorized):
		return writeError(c, fiber.StatusUnauthorized, cleanServiceError(err, service.ErrUnauthorized))
	case errors.Is(err, service.ErrInvalidGrant):
		return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidGrant))
	case errors.Is(err, service.ErrInvalidClient):
		return writeError(c, fiber.StatusUnauthorized, cleanServiceError(err, service.ErrInvalidClient))
	case errors.Is(err, service.ErrInvalidToken):
		return writeError(c, fiber.StatusUnauthorized, cleanServiceError(err, service.ErrInvalidToken))
	case errors.Is(err, service.ErrAccessDenied):
		return writeError(c, fiber.StatusForbidden, cleanServiceError(err, service.ErrAccessDenied))
	default:
		return writeError(c, fiber.StatusInternalServerError, fallback)
	}
}

func cleanServiceError(err error, base error) string {
	prefix := base.Error() + ": "
	message := err.Error()
	if strings.HasPrefix(message, prefix) {
		return strings.TrimPrefix(message, prefix)
	}
	return message
}
