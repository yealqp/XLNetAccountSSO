package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/xianlin-network/sso-platform/server/internal/http/middleware"
	"github.com/xianlin-network/sso-platform/server/internal/service"
)

type AdminHandler struct {
	adminService *service.AdminService
	tokenService *service.TokenService
}

func NewAdminHandler(adminService *service.AdminService, tokenService *service.TokenService) *AdminHandler {
	return &AdminHandler{adminService: adminService, tokenService: tokenService}
}

func (handler *AdminHandler) Overview(c *fiber.Ctx) error {
	overview, err := handler.adminService.Overview(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load overview failed")
	}
	return c.JSON(overview)
}

func (handler *AdminHandler) PublicSettings(c *fiber.Ctx) error {
	settings, err := handler.adminService.PublicSettings(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load public settings failed")
	}
	return c.JSON(settings)
}

func (handler *AdminHandler) Me(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	return c.JSON(fiber.Map{"user": publicUser(authContext.User)})
}

func (handler *AdminHandler) PlatformSettings(c *fiber.Ctx) error {
	settings, err := handler.adminService.PublicSettings(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load platform settings failed")
	}
	return c.JSON(settings)
}

func (handler *AdminHandler) UpdatePlatformSettings(c *fiber.Ctx) error {
	var input struct {
		PlatformName string `json:"platform_name"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid settings payload")
	}
	settings, err := handler.adminService.UpdatePlatformName(context.Background(), input.PlatformName)
	if err != nil {
		return handleServiceError(c, err, "update platform settings failed")
	}
	return c.JSON(settings)
}

func (handler *AdminHandler) ListClients(c *fiber.Ctx) error {
	clients, err := handler.adminService.ListClients(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load clients failed")
	}
	return c.JSON(fiber.Map{"items": clients})
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
	return c.Status(fiber.StatusCreated).JSON(client)
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
	return c.Status(fiber.StatusCreated).JSON(result)
}

func (handler *AdminHandler) UpdateClient(c *fiber.Ctx) error {
	var input service.UpdateClientInput
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid client payload")
	}
	client, err := handler.adminService.UpdateClient(context.Background(), c.Params("id"), input)
	if err != nil {
		return handleServiceError(c, err, "update client failed")
	}
	return c.JSON(client)
}

func (handler *AdminHandler) DeleteClient(c *fiber.Ctx) error {
	err := handler.adminService.DeleteClient(context.Background(), c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "delete client failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *AdminHandler) ListUsers(c *fiber.Ctx) error {
	users, err := handler.adminService.ListUsers(context.Background())
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load users failed")
	}
	return c.JSON(fiber.Map{"items": users})
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
	return c.Status(fiber.StatusCreated).JSON(user)
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
	return c.JSON(user)
}

func (handler *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.adminService.DeleteUser(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "delete user failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *AdminHandler) ListTokens(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	tokens, err := handler.tokenService.ListTokens(context.Background(), authContext.User)
	if err != nil {
		return writeError(c, fiber.StatusInternalServerError, "load tokens failed")
	}
	return c.JSON(fiber.Map{"items": tokens})
}

func (handler *AdminHandler) RevokeAccessToken(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.tokenService.RevokeAccessToken(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "revoke access token failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *AdminHandler) RevokeRefreshToken(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.tokenService.RevokeRefreshToken(context.Background(), authContext.User, c.Params("id"))
	if err != nil {
		return handleServiceError(c, err, "revoke refresh token failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *AdminHandler) RevokeClientTokens(c *fiber.Ctx) error {
	authContext, _ := middleware.CurrentAuthContext(c)
	err := handler.tokenService.RevokeClientTokens(context.Background(), authContext.User, c.Params("clientId"))
	if err != nil {
		return handleServiceError(c, err, "revoke client tokens failed")
	}
	return c.SendStatus(fiber.StatusNoContent)
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
