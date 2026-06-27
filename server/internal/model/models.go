package model

import (
	"strings"
	"time"
)

type User struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username           string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash       string    `gorm:"size:255;not null" json:"-"`
	Email              string    `gorm:"size:160;not null" json:"email"`
	Role               string    `gorm:"size:32;not null;default:user" json:"role"`
	Status             string    `gorm:"size:32;not null;default:active" json:"status"`
	WebAuthnUserHandle *string   `gorm:"size:128;uniqueIndex" json:"-"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type OAuthClient struct {
	ID               string    `gorm:"primaryKey;size:36" json:"id"`
	Name             string    `gorm:"size:160;not null" json:"name"`
	Description      string    `gorm:"type:text" json:"description"`
	IconURL          string    `gorm:"size:500" json:"icon_url"`
	ClientID         string    `gorm:"size:120;uniqueIndex;not null" json:"client_id"`
	ClientSecretHash string    `gorm:"size:255" json:"-"`
	ClientType       string    `gorm:"size:32;not null;default:public" json:"client_type"`
	RedirectURIsRaw  string    `gorm:"type:text;not null" json:"-"`
	ScopesRaw        string    `gorm:"type:text;not null" json:"-"`
	Trusted          bool      `gorm:"not null;default:false" json:"trusted"`
	CreatedBy        uint      `gorm:"not null" json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type AuthorizationCode struct {
	ID                  string     `gorm:"primaryKey;size:36" json:"id"`
	CodeHash            string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	ClientID            string     `gorm:"size:120;index;not null" json:"client_id"`
	UserID              uint       `gorm:"index;not null" json:"user_id"`
	RedirectURI         string     `gorm:"size:500;not null" json:"redirect_uri"`
	Scope               string     `gorm:"type:text;not null" json:"scope"`
	State               string     `gorm:"size:255;not null" json:"state"`
	Nonce               string     `gorm:"size:255" json:"nonce"`
	CodeChallenge       string     `gorm:"size:255;not null" json:"-"`
	CodeChallengeMethod string     `gorm:"size:32;not null" json:"-"`
	ExpiresAt           time.Time  `gorm:"index;not null" json:"expires_at"`
	ConsumedAt          *time.Time `json:"consumed_at"`
	CreatedAt           time.Time  `json:"created_at"`
}

type AccessToken struct {
	ID        string     `gorm:"primaryKey;size:36" json:"id"`
	TokenHash string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	ClientID  string     `gorm:"size:120;index;not null" json:"client_id"`
	UserID    uint       `gorm:"index;not null" json:"user_id"`
	Scope     string     `gorm:"type:text;not null" json:"scope"`
	ExpiresAt time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type RefreshToken struct {
	ID            string     `gorm:"primaryKey;size:36" json:"id"`
	TokenHash     string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	AccessTokenID string     `gorm:"size:36;index;not null" json:"access_token_id"`
	ClientID      string     `gorm:"size:120;index;not null" json:"client_id"`
	UserID        uint       `gorm:"index;not null" json:"user_id"`
	Scope         string     `gorm:"type:text;not null" json:"scope"`
	ExpiresAt     time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt     *time.Time `json:"revoked_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type UserSession struct {
	ID               string     `gorm:"primaryKey;size:36" json:"id"`
	SessionTokenHash string     `gorm:"size:128;uniqueIndex;not null" json:"-"`
	UserID           uint       `gorm:"index;not null" json:"user_id"`
	IPAddress        string     `gorm:"size:64" json:"ip_address"`
	UserAgent        string     `gorm:"size:255" json:"user_agent"`
	LastSeenAt       time.Time  `gorm:"not null" json:"last_seen_at"`
	ExpiresAt        time.Time  `gorm:"index;not null" json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type UserPasskeyCredential struct {
	ID             string     `gorm:"primaryKey;size:36" json:"id"`
	UserID         uint       `gorm:"index;not null" json:"user_id"`
	Name           string     `gorm:"size:160;not null" json:"name"`
	CredentialID   string     `gorm:"size:512;uniqueIndex;not null" json:"-"`
	CredentialJSON string     `gorm:"type:longtext;not null" json:"-"`
	LastUsedAt     *time.Time `json:"last_used_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type WebAuthnCeremony struct {
	ID          string     `gorm:"primaryKey;size:64" json:"id"`
	Purpose     string     `gorm:"size:32;index;not null" json:"purpose"`
	UserID      *uint      `gorm:"index" json:"-"`
	SessionData string     `gorm:"type:longtext;not null" json:"-"`
	ExpiresAt   time.Time  `gorm:"index;not null" json:"expires_at"`
	ConsumedAt  *time.Time `json:"consumed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type AuditLog struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	ActorID   uint      `gorm:"index" json:"actor_id"`
	Action    string    `gorm:"size:120;index;not null" json:"action"`
	Target    string    `gorm:"size:255;not null" json:"target"`
	IPAddress string    `gorm:"size:64" json:"ip_address"`
	Metadata  string    `gorm:"type:text" json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

type PlatformSetting struct {
	Key       string    `gorm:"primaryKey;size:80" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Asset struct {
	Key      string    `gorm:"primaryKey;size:120" json:"key"`
	Data     string    `gorm:"type:longtext;not null" json:"data"`
	MimeType string    `gorm:"size:80;not null" json:"mime_type"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type EmailVerificationCode struct {
	ID         string     `gorm:"primaryKey;size:36" json:"id"`
	Email      string     `gorm:"size:160;index;not null" json:"email"`
	Purpose    string     `gorm:"size:32;index;not null" json:"purpose"`
	CodeHash   string     `gorm:"size:128;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"index;not null" json:"expires_at"`
	ConsumedAt *time.Time `json:"consumed_at"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (client OAuthClient) RedirectURIs() []string {
	return splitLines(client.RedirectURIsRaw)
}

func (client *OAuthClient) SetRedirectURIs(values []string) {
	client.RedirectURIsRaw = strings.Join(normalizeList(values), "\n")
}

func (client OAuthClient) Scopes() []string {
	return splitScopes(client.ScopesRaw)
}

func (client *OAuthClient) SetScopes(values []string) {
	client.ScopesRaw = strings.Join(normalizeList(values), " ")
}

func splitLines(value string) []string {
	parts := strings.Split(value, "\n")
	return normalizeList(parts)
}

func splitScopes(value string) []string {
	value = strings.ReplaceAll(value, ",", " ")
	parts := strings.Fields(value)
	return normalizeList(parts)
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
