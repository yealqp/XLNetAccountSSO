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
	authService := service.NewAuthService(store, cfg)
	if err := authService.SeedDefaults(context.Background()); err != nil {
		return nil, fmt.Errorf("seed defaults: %w", err)
	}
	adminService := service.NewAdminService(store, cfg)
	if err := adminService.EnsureDefaults(context.Background()); err != nil {
		return nil, fmt.Errorf("seed platform settings: %w", err)
	}
	oauthService, err := service.NewOAuthService(store, cfg)
	if err != nil {
		return nil, fmt.Errorf("build oidc service: %w", err)
	}
	tokenService := service.NewTokenService(store)

	authHandler := handlers.NewAuthHandler(authService, cfg)
	adminHandler := handlers.NewAdminHandler(adminService, tokenService)
	oauthHandler := handlers.NewOAuthHandler(oauthService, cfg)

	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     joinOrigins(cfg.AllowedOrigins),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))
	app.Use(middleware.OptionalSession(authService, cfg.CookieName))

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
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/logout", authHandler.Logout)
	api.Get("/auth/session", authHandler.Session)
	api.Get("/settings/public", adminHandler.PublicSettings)

	secured := api.Group("", middleware.RequireSession(authService, cfg.CookieName))
	secured.Get("/me", adminHandler.Me)
	secured.Get("/oauth/requests/preview", oauthHandler.Preview)
	secured.Post("/oauth/requests/decision", oauthHandler.Decide)
	secured.Get("/tokens", adminHandler.ListTokens)
	secured.Delete("/tokens/access/:id", adminHandler.RevokeAccessToken)
	secured.Delete("/tokens/refresh/:id", adminHandler.RevokeRefreshToken)
	secured.Delete("/tokens/client/:clientId", adminHandler.RevokeClientTokens)

	admin := secured.Group("", middleware.RequireAdmin())
	admin.Get("/overview", adminHandler.Overview)
	admin.Get("/settings/platform", adminHandler.PlatformSettings)
	admin.Put("/settings/platform", adminHandler.UpdatePlatformSettings)
	admin.Post("/client-icons/upload", adminHandler.UploadClientIcon)
	admin.Get("/clients", adminHandler.ListClients)
	admin.Post("/clients", adminHandler.CreateClient)
	admin.Put("/clients/:id", adminHandler.UpdateClient)
	admin.Delete("/clients/:id", adminHandler.DeleteClient)
	admin.Get("/users", adminHandler.ListUsers)
	admin.Post("/users", adminHandler.CreateUser)
	admin.Put("/users/:id", adminHandler.UpdateUser)
	admin.Delete("/users/:id", adminHandler.DeleteUser)

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
