package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/http/handlers"
	"github.com/XianLinNet/XLNetAccount/internal/http/middleware"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
)

func Build(cfg config.Config, db *gorm.DB) (*fiber.App, error) {
	store := repository.New(db)
	authService := service.NewAuthService(store)
	if err := authService.SeedDefaults(context.Background()); err != nil {
		return nil, fmt.Errorf("seed defaults: %w", err)
	}
	adminService := service.NewAdminService(store, cfg)
	if err := adminService.EnsureDefaults(context.Background()); err != nil {
		return nil, fmt.Errorf("seed platform settings: %w", err)
	}
	verificationService := service.NewVerificationService(store)
	passkeyService, err := service.NewPasskeyService(store, cfg, authService)
	if err != nil {
		return nil, fmt.Errorf("build passkey service: %w", err)
	}
	totpService := service.NewTOTPService(store, cfg, authService)
	oauthService, err := service.NewOAuthService(store, cfg)
	if err != nil {
		return nil, fmt.Errorf("build oidc service: %w", err)
	}
	tokenService := service.NewTokenService(store)

	authHandler := handlers.NewAuthHandler(authService, passkeyService, adminService, verificationService, totpService, cfg)
	adminHandler := handlers.NewAdminHandler(adminService, tokenService)
	oauthHandler := handlers.NewOAuthHandler(oauthService, cfg)

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${method} ${path} ${status} ${latency} ${bytesSent}B\n",
		Output: nil,
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: joinOrigins(cfg.AllowedOrigins),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,OPTIONS",
	}))
	app.Use(middleware.OptionalSession(authService))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	apiAssets := app.Group("/api/assets")
	apiAssets.Get("/*", func(c *fiber.Ctx) error {
		key := strings.TrimPrefix(c.Params("*"), "/")
		if key == "" {
			return c.Status(404).JSON(fiber.Map{"code": 404, "message": "not found"})
		}
		asset, err := store.FindAsset(c.Context(), key)
		if err != nil || asset == nil {
			return c.Status(404).JSON(fiber.Map{"code": 404, "message": "not found"})
		}
		data, err := base64.StdEncoding.DecodeString(asset.Data)
		if err != nil {
			slog.Error("decode asset", "key", key, "error", err)
			return c.Status(500).JSON(fiber.Map{"code": 500, "message": "decode failed"})
		}
		c.Type(asset.MimeType)
		return c.Send(data)
	})

	app.Get("/.well-known/openid-configuration", oauthHandler.OpenIDConfiguration)
	app.Get("/.well-known/jwks.json", oauthHandler.JWKS)
	app.Get("/oauth/authorize", oauthHandler.AuthorizeEntry)
	app.Post("/oauth/token", oauthHandler.Token)
	app.Post("/oauth/revoke", oauthHandler.Revoke)
	app.Post("/oauth/introspect", oauthHandler.Introspect)
	app.Get("/oauth/userinfo", oauthHandler.UserInfo)

	api := app.Group("/api")
	api.Get("/setup/status", authHandler.SetupStatus)
	api.Post("/setup/initialize", authHandler.Initialize)
	api.Post("/auth/register/code/send", authHandler.SendRegisterCode)
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/passkeys/login/start", authHandler.BeginPasskeyLogin)
	api.Post("/auth/passkeys/login/finish", authHandler.FinishPasskeyLogin)
	api.Post("/auth/logout", authHandler.Logout)
	api.Post("/auth/totp/login/verify", authHandler.VerifyTOTPLogin)
	api.Get("/auth/session", authHandler.Session)
	api.Get("/settings/public", adminHandler.PublicSettings)

	secured := api.Group("", middleware.RequireSession(authService))
	secured.Get("/me", adminHandler.Me)
	secured.Post("/me/password/code/send", authHandler.SendProfilePasswordCode)
	secured.Post("/me/profile", authHandler.UpdateProfile)
	secured.Get("/me/passkeys", authHandler.ListPasskeys)
	secured.Post("/me/passkeys/register/start", authHandler.BeginPasskeyRegistration)
	secured.Post("/me/passkeys/register/finish", authHandler.FinishPasskeyRegistration)
	secured.Post("/me/passkeys/:id/delete", authHandler.DeletePasskey)
	secured.Get("/me/totp/status", authHandler.TOTPStatus)
	secured.Post("/me/totp/setup/start", authHandler.BeginTOTPSetup)
	secured.Post("/me/totp/setup/verify", authHandler.VerifyTOTPSetup)
	secured.Post("/me/totp/disable", authHandler.DisableTOTP)
	secured.Get("/overview", adminHandler.Overview)
	secured.Post("/client-icons/upload", adminHandler.UploadClientIcon)
	secured.Get("/oauth/requests/preview", oauthHandler.Preview)
	secured.Post("/oauth/requests/decision", oauthHandler.Decide)
	secured.Get("/clients", adminHandler.ListClients)
	secured.Post("/clients", adminHandler.CreateClient)
	secured.Post("/clients/:id/update", adminHandler.UpdateClient)
	secured.Post("/clients/:id/delete", adminHandler.DeleteClient)
	secured.Get("/tokens", adminHandler.ListTokens)
	secured.Post("/tokens/access/:id/revoke", adminHandler.RevokeAccessToken)
	secured.Post("/tokens/refresh/:id/revoke", adminHandler.RevokeRefreshToken)
	secured.Post("/tokens/client/:clientId/revoke", adminHandler.RevokeClientTokens)

	admin := secured.Group("", middleware.RequireAdmin())
	admin.Get("/manage/overview", adminHandler.AdminOverview)
	admin.Get("/settings/platform", adminHandler.PlatformSettings)
	admin.Post("/settings/platform", adminHandler.UpdatePlatformSettings)
	admin.Post("/settings/platform/icon/upload", adminHandler.UploadWebIcon)
	admin.Post("/settings/platform/test-email", adminHandler.SendTestEmail)
	admin.Get("/manage/clients", adminHandler.ListManagedClients)
	admin.Post("/manage/clients/:id/update", adminHandler.UpdateManagedClient)
	admin.Post("/manage/clients/:id/delete", adminHandler.DeleteManagedClient)
	admin.Get("/users", adminHandler.ListUsers)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Post("/users/:id/update", adminHandler.UpdateUser)
	admin.Post("/users/:id/delete", adminHandler.DeleteUser)
	admin.Get("/manage/tokens", adminHandler.ListManagedTokens)

	return app, nil
}

func joinOrigins(origins []string) string {
	result := ""
	for index, origin := range origins {
		if index > 0 {
			result += ","
		}
		result += origin
	}
	return result
}
