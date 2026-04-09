package handlers

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/XianLinNet/XLNetAccount/internal/http/middleware"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

func (handler *AuthHandler) BeginPasskeyRegistration(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	sessionID, options, err := handler.passkeyService.BeginRegistration(context.Background(), authContext.User)
	if err != nil {
		return handlePasskeyError(c, err, "start passkey registration failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{
		"session_id": sessionID,
		"options":    options,
	}, "success")
}

func (handler *AuthHandler) FinishPasskeyRegistration(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	var input struct {
		SessionID  string          `json:"session_id"`
		Name       string          `json:"name"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid passkey registration payload")
	}
	credential, err := handler.passkeyService.FinishRegistration(context.Background(), authContext.User, input.SessionID, input.Name, input.Credential)
	if err != nil {
		return handlePasskeyError(c, err, "finish passkey registration failed")
	}
	return writeSuccess(c, fiber.StatusCreated, fiber.Map{"credential": credential}, "success")
}

func (handler *AuthHandler) ListPasskeys(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	items, err := handler.passkeyService.ListPasskeys(context.Background(), authContext.User)
	if err != nil {
		return handlePasskeyError(c, err, "load passkeys failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{"items": items}, "success")
}

func (handler *AuthHandler) DeletePasskey(c *fiber.Ctx) error {
	authContext, err := middleware.CurrentAuthContext(c)
	if err != nil {
		return writeError(c, fiber.StatusUnauthorized, "authentication required")
	}
	if err := handler.passkeyService.DeletePasskey(context.Background(), authContext.User, c.Params("id")); err != nil {
		return handlePasskeyError(c, err, "delete passkey failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{}, "success")
}

func (handler *AuthHandler) BeginPasskeyLogin(c *fiber.Ctx) error {
	sessionID, options, err := handler.passkeyService.BeginLogin(context.Background())
	if err != nil {
		return handlePasskeyError(c, err, "start passkey login failed")
	}
	return writeSuccess(c, fiber.StatusOK, fiber.Map{
		"session_id": sessionID,
		"options":    options,
	}, "success")
}

func (handler *AuthHandler) FinishPasskeyLogin(c *fiber.Ctx) error {
	var input struct {
		SessionID  string          `json:"session_id"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := c.BodyParser(&input); err != nil {
		return writeError(c, fiber.StatusBadRequest, "invalid passkey login payload")
	}
	user, rawSessionToken, session, err := handler.passkeyService.FinishLogin(context.Background(), input.SessionID, input.Credential, service.SessionMeta{
		IPAddress: c.IP(),
		UserAgent: c.Get(fiber.HeaderUserAgent),
	})
	if err != nil {
		return handlePasskeyError(c, err, "finish passkey login failed")
	}
	return writeSuccess(c, fiber.StatusOK, authSessionPayload(user, rawSessionToken, session), "success")
}

func handlePasskeyError(c *fiber.Ctx, err error, fallback string) error {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		return writeError(c, fiber.StatusUnauthorized, "passkey authentication failed")
	case errors.Is(err, service.ErrForbidden):
		return writeError(c, fiber.StatusForbidden, cleanServiceError(err, service.ErrForbidden))
	case errors.Is(err, service.ErrConflict):
		return writeError(c, fiber.StatusConflict, cleanServiceError(err, service.ErrConflict))
	case errors.Is(err, service.ErrInvalidInput):
		return writeError(c, fiber.StatusBadRequest, cleanServiceError(err, service.ErrInvalidInput))
	case errors.Is(err, service.ErrInvalidGrant):
		return writeError(c, fiber.StatusBadRequest, "passkey session is invalid or expired")
	case errors.Is(err, service.ErrNotFound):
		return writeError(c, fiber.StatusNotFound, "passkey not found")
	default:
		return writeError(c, fiber.StatusInternalServerError, fallback)
	}
}
