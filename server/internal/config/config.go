package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName                string
	Port                   string
	ServerBaseURL          string
	WebBaseURL             string
	AssetDir               string
	OIDCIssuer             string
	OIDCKeyID              string
	OIDCPrivateKeyPEM      string
	AllowedOrigins         []string
	CookieName             string
	CookieSecure           bool
	DBDSN                  string
	SeedAdminUsername      string
	SeedAdminPassword      string
	SeedAdminDisplayName   string
	SeedAdminEmail         string
	SeedDemoClientName     string
	SeedDemoClientID       string
	SeedDemoClientRedirect string
}

func Load() Config {
	dsn := getEnv("DB_DSN", "")
	if dsn == "" {
		host := getEnv("DB_HOST", "127.0.0.1")
		port := getEnv("DB_PORT", "3306")
		user := getEnv("DB_USER", "root")
		password := getEnv("DB_PASSWORD", "root")
		name := getEnv("DB_NAME", "sso_platform")
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)
	}

	serverBaseURL := strings.TrimRight(getEnv("SERVER_BASE_URL", "http://localhost:8080"), "/")
	webBaseURL := strings.TrimRight(getEnv("WEB_BASE_URL", "http://localhost:5173"), "/")

	return Config{
		AppName:                getEnv("APP_NAME", "XLNetAccount"),
		Port:                   getEnv("PORT", "8080"),
		ServerBaseURL:          serverBaseURL,
		WebBaseURL:             webBaseURL,
		AssetDir:               getEnv("ASSET_DIR", "./data"),
		OIDCIssuer:             strings.TrimRight(getEnv("OIDC_ISSUER", serverBaseURL), "/"),
		OIDCKeyID:              getEnv("OIDC_KEY_ID", ""),
		OIDCPrivateKeyPEM:      getEnv("OIDC_PRIVATE_KEY_PEM", ""),
		AllowedOrigins:         splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:4173")),
		CookieName:             getEnv("COOKIE_NAME", "sso_session"),
		CookieSecure:           parseBool(getEnv("COOKIE_SECURE", "false")),
		DBDSN:                  dsn,
		SeedAdminUsername:      getEnv("SEED_ADMIN_USERNAME", "admin"),
		SeedAdminPassword:      getEnv("SEED_ADMIN_PASSWORD", "Admin123!"),
		SeedAdminDisplayName:   getEnv("SEED_ADMIN_DISPLAY_NAME", "Platform Admin"),
		SeedAdminEmail:         getEnv("SEED_ADMIN_EMAIL", "admin@example.com"),
		SeedDemoClientName:     getEnv("SEED_DEMO_CLIENT_NAME", "Demo Client"),
		SeedDemoClientID:       getEnv("SEED_DEMO_CLIENT_ID", "demo-web-client"),
		SeedDemoClientRedirect: getEnv("SEED_DEMO_CLIENT_REDIRECT_URI", "http://localhost:4173/callback"),
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseBool(value string) bool {
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}
	return parsed
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}
