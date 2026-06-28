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
	cfg := config.Load()

	dbLogger := logger.New(
		slog.NewLogLogger(slog.Default().Handler(), slog.LevelWarn),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		},
	)

	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{
		Logger: dbLogger,
	})
	if err != nil {
		slog.Error("connect database", "error", err)
		os.Exit(1)
	}

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
		&model.UserTOTP{},
		&model.AuditLog{},
		&model.Asset{},
	); err != nil {
		slog.Error("auto migrate schema", "error", err)
		os.Exit(1)
	}

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
