package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/app"
	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	slog.Info("starting server", "port_env", os.Getenv("PORT"))

	cfg := config.Load()
	slog.Info("config loaded", "port", cfg.Port, "db_host", os.Getenv("DB_HOST"))

	dbLogger := logger.New(
		slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		},
	)

	slog.Info("connecting to database")
	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	slog.Info("running auto migrate")
	if err := db.AutoMigrate(
		&model.User{},
		&model.OAuthClient{},
		&model.PlatformSetting{},
		&model.EmailVerificationCode{},
		&model.AuthorizationCode{},
		&model.AccessToken{},
		&model.RefreshToken{},
		&model.UserSession{},
		&model.UserPasskeyCredential{},
		&model.WebAuthnCeremony{},
		&model.AuditLog{},
		&model.Asset{},
	); err != nil {
		slog.Error("auto migrate schema", "error", err)
		os.Exit(1)
	}
	slog.Info("auto migrate complete")

	server, err := app.Build(cfg, db)
	if err != nil {
		slog.Error("build app", "error", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	slog.Info("server starting", "app", cfg.AppName, "addr", addr)
	if err := server.Listen(addr); err != nil {
		slog.Error("start server", "error", err)
		os.Exit(1)
	}
}
