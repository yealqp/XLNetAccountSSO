package main

import (
	"fmt"
	"log"

	"github.com/XianLinNet/XLNetAccount/internal/app"
	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(mysql.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
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
		&model.AuditLog{},
	); err != nil {
		log.Fatalf("auto migrate schema: %v", err)
	}

	server, err := app.Build(cfg, db)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("%s listening on %s", cfg.AppName, addr)
	if err := server.Listen(addr); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
