package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppName           string
	Port              string
	ServerBaseURL     string
	WebBaseURL        string
	WebAuthnRPID      string
	WebAuthnRPOrigins []string
	OIDCIssuer        string
	OIDCKeyID         string
	OIDCPrivateKeyPEM string
	AllowedOrigins    []string
	DBDSN             string

	SMTPHost     string
	SMTPUser     string
	SMTPPassword string
	SMTPPort     string
	SMTPTLS      bool

	CAPAPIEndpoint string
	CAPSiteKey     string
	CAPSecretKey   string
}

func Load() Config {
	host := getEnv("DB_HOST", "127.0.0.1")
	port := getEnv("DB_PORT", "3306")
	user := getEnv("DB_USER", "root")
	password := getEnv("DB_PWD", "root")
	name := getEnv("DB_NAME", "sso_platform")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, password, host, port, name)

	serverBaseURL := strings.TrimRight(getEnv("SERVER_BASE_URL", "http://localhost:8080"), "/")
	webBaseURL := strings.TrimRight(getEnv("WEB_BASE_URL", serverBaseURL), "/")
	allowedOrigins := splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173"))
	webAuthnRPOrigins := normalizeOrigins(splitCSV(getEnv("WEBAUTHN_RP_ORIGINS", "")))
	if len(webAuthnRPOrigins) == 0 {
		webAuthnRPOrigins = defaultWebAuthnOrigins(webBaseURL, serverBaseURL, allowedOrigins)
	}
	webAuthnRPID := strings.TrimSpace(getEnv("WEBAUTHN_RP_ID", ""))
	if webAuthnRPID == "" {
		webAuthnRPID = firstAvailableHost(webBaseURL, allowedOrigins, serverBaseURL)
	}

	return Config{
		AppName:           getEnv("APP_NAME", "XLNetAccount"),
		Port:              getEnv("PORT", "9000"),
		ServerBaseURL:     serverBaseURL,
		WebBaseURL:        webBaseURL,
		WebAuthnRPID:      webAuthnRPID,
		WebAuthnRPOrigins: webAuthnRPOrigins,
		OIDCIssuer:        strings.TrimRight(getEnv("OIDC_ISSUER", serverBaseURL), "/"),
		OIDCKeyID:         getEnv("OIDC_KEY_ID", ""),
		OIDCPrivateKeyPEM: getEnv("OIDC_PRIVATE_KEY_PEM", ""),
		AllowedOrigins:    allowedOrigins,
		DBDSN:             dsn,

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PWD", ""),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPTLS:      parseEnvBool("SMTP_TLS", true),

		CAPAPIEndpoint: getEnv("CAP_API_ENDPOINT", ""),
		CAPSiteKey:     getEnv("CAP_SITE_KEY", ""),
		CAPSecretKey:   getEnv("CAP_SECRET_KEY", ""),
	}
}

func getEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
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

func normalizeOrigins(values []string) []string {
	items := normalizeList(values)
	origins := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimRight(strings.TrimSpace(item), "/")
		if trimmed == "" {
			continue
		}
		origins = append(origins, trimmed)
	}
	return origins
}

func defaultWebAuthnOrigins(webBaseURL string, serverBaseURL string, allowedOrigins []string) []string {
	origins := normalizeOrigins(append([]string{webBaseURL, serverBaseURL}, allowedOrigins...))
	if len(origins) > 0 {
		return origins
	}
	return []string{"http://localhost:8080"}
}

func hostFromURL(raw string) string {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Hostname())
}

func firstAvailableHost(primary string, alternatives []string, fallback string) string {
	for _, candidate := range append([]string{primary}, append(alternatives, fallback)...) {
		host := hostFromURL(candidate)
		if host != "" {
			return host
		}
	}
	return "localhost"
}

func normalizeList(values []string) []string {
	items := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		items = append(items, trimmed)
	}
	return items
}

func (cfg Config) BaseURL() string {
	baseURL := strings.TrimSpace(cfg.ServerBaseURL)
	if baseURL == "" {
		baseURL = "http://localhost"
		if strings.TrimSpace(cfg.Port) != "" && strings.TrimSpace(cfg.Port) != "80" {
			baseURL += ":" + strings.TrimSpace(cfg.Port)
		}
	}
	return strings.TrimRight(baseURL, "/")
}

func (cfg Config) IssuerURL() string {
	issuer := strings.TrimSpace(cfg.OIDCIssuer)
	if issuer == "" {
		return cfg.BaseURL()
	}
	return strings.TrimRight(issuer, "/")
}

func (cfg Config) PublicURL(path string) string {
	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return cfg.BaseURL()
	}
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}
	return cfg.BaseURL() + trimmedPath
}
