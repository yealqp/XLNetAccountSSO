package app

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/xianlin-network/sso-platform/server/internal/config"
	"github.com/xianlin-network/sso-platform/server/internal/http/handlers"
	"github.com/xianlin-network/sso-platform/server/internal/http/middleware"
	"github.com/xianlin-network/sso-platform/server/internal/repository"
	"github.com/xianlin-network/sso-platform/server/internal/service"
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
	oauthService, err := service.NewOAuthService(store, cfg)
	if err != nil {
		return nil, fmt.Errorf("build oidc service: %w", err)
	}
	tokenService := service.NewTokenService(store)

	authHandler := handlers.NewAuthHandler(authService, adminService, verificationService)
	adminHandler := handlers.NewAdminHandler(adminService, tokenService)
	oauthHandler := handlers.NewOAuthHandler(oauthService, cfg)

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: joinOrigins(cfg.AllowedOrigins),
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))
	app.Use(middleware.OptionalSession(authService))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Static("/client-icons", service.ClientIconDir(cfg))

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
	api.Post("/auth/logout", authHandler.Logout)
	api.Get("/auth/session", authHandler.Session)
	api.Get("/settings/public", adminHandler.PublicSettings)

	secured := api.Group("", middleware.RequireSession(authService))
	secured.Get("/me", adminHandler.Me)
	secured.Post("/me/password/code/send", authHandler.SendProfilePasswordCode)
	secured.Put("/me/profile", authHandler.UpdateProfile)
	secured.Get("/overview", adminHandler.Overview)
	secured.Post("/client-icons/upload", adminHandler.UploadClientIcon)
	secured.Get("/oauth/requests/preview", oauthHandler.Preview)
	secured.Post("/oauth/requests/decision", oauthHandler.Decide)
	secured.Get("/clients", adminHandler.ListClients)
	secured.Post("/clients", adminHandler.CreateClient)
	secured.Put("/clients/:id", adminHandler.UpdateClient)
	secured.Delete("/clients/:id", adminHandler.DeleteClient)
	secured.Get("/tokens", adminHandler.ListTokens)
	secured.Delete("/tokens/access/:id", adminHandler.RevokeAccessToken)
	secured.Delete("/tokens/refresh/:id", adminHandler.RevokeRefreshToken)
	secured.Delete("/tokens/client/:clientId", adminHandler.RevokeClientTokens)

	admin := secured.Group("", middleware.RequireAdmin())
	admin.Get("/manage/overview", adminHandler.AdminOverview)
	admin.Get("/settings/platform", adminHandler.PlatformSettings)
	admin.Put("/settings/platform", adminHandler.UpdatePlatformSettings)
	admin.Post("/settings/platform/test-email", adminHandler.SendTestEmail)
	admin.Get("/manage/clients", adminHandler.ListManagedClients)
	admin.Put("/manage/clients/:id", adminHandler.UpdateManagedClient)
	admin.Delete("/manage/clients/:id", adminHandler.DeleteManagedClient)
	admin.Get("/users", adminHandler.ListUsers)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Put("/users/:id", adminHandler.UpdateUser)
	admin.Delete("/users/:id", adminHandler.DeleteUser)
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
