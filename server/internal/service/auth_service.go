package service

import (
	"context"
	"strings"
	"time"

	"github.com/xianlin-network/sso-platform/server/internal/model"
	"github.com/xianlin-network/sso-platform/server/internal/pkg/security"
	"github.com/xianlin-network/sso-platform/server/internal/repository"
)

const sessionLifetime = 7 * 24 * time.Hour

type SessionMeta struct {
	IPAddress string
	UserAgent string
}

type AuthService struct {
	store *repository.Store
}

func NewAuthService(store *repository.Store) *AuthService {
	return &AuthService{store: store}
}

func (service *AuthService) SeedDefaults(ctx context.Context) error {
	return nil
}

type InitializeAdminInput struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type RegisterInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Code        string `json:"code"`
}

func (service *AuthService) IsInitialized(ctx context.Context) (bool, error) {
	count, err := service.store.CountUsers(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (service *AuthService) InitializeFirstAdmin(ctx context.Context, input InitializeAdminInput) (*model.User, error) {
	initialized, err := service.IsInitialized(ctx)
	if err != nil {
		return nil, err
	}
	if initialized {
		return nil, ErrConflict
	}
	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return nil, ErrInvalidInput
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		ID:           security.NewID(),
		Username:     username,
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(input.DisplayName),
		Email:        strings.TrimSpace(strings.ToLower(input.Email)),
		Role:         "admin",
		Status:       "active",
	}
	if user.DisplayName == "" {
		user.DisplayName = username
	}
	if err := service.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	service.logAudit(ctx, user.ID, "setup.initialize", user.Username, "", "")
	return user, nil
}

func (service *AuthService) Login(ctx context.Context, username string, password string, meta SessionMeta) (*model.User, string, *model.UserSession, error) {
	initialized, err := service.IsInitialized(ctx)
	if err != nil {
		return nil, "", nil, err
	}
	if !initialized {
		return nil, "", nil, ErrConflict
	}
	user, err := service.findUserByIdentifier(ctx, username)
	if err != nil {
		return nil, "", nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, "", nil, ErrUnauthorized
	}
	if err := security.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, "", nil, ErrUnauthorized
	}

	rawSessionToken, err := security.NewOpaqueToken(32)
	if err != nil {
		return nil, "", nil, err
	}

	now := time.Now().UTC()
	session := &model.UserSession{
		ID:               security.NewID(),
		SessionTokenHash: security.HashToken(rawSessionToken),
		UserID:           user.ID,
		IPAddress:        meta.IPAddress,
		UserAgent:        trimLength(meta.UserAgent, 255),
		LastSeenAt:       now,
		ExpiresAt:        now.Add(sessionLifetime),
	}

	if err := service.store.CreateSession(ctx, session); err != nil {
		return nil, "", nil, err
	}

	service.logAudit(ctx, user.ID, "auth.login", user.Username, meta.IPAddress, "")
	return user, rawSessionToken, session, nil
}

func (service *AuthService) Register(ctx context.Context, input RegisterInput) (*model.User, error) {
	initialized, err := service.IsInitialized(ctx)
	if err != nil {
		return nil, err
	}
	if !initialized {
		return nil, ErrConflict
	}
	allowed, err := service.registrationEnabled(ctx)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrForbidden
	}
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := strings.TrimSpace(input.Password)
	if email == "" || password == "" || strings.TrimSpace(input.Code) == "" || !strings.Contains(email, "@") {
		return nil, ErrInvalidInput
	}
	existingByEmail, err := service.store.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingByEmail != nil {
		return nil, ErrConflict
	}
	existingByUsername, err := service.store.FindUserByUsername(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingByUsername != nil {
		return nil, ErrConflict
	}
	verificationService := NewVerificationService(service.store)
	if err := verificationService.VerifyRegistrationCode(ctx, email, input.Code); err != nil {
		return nil, err
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		ID:           security.NewID(),
		Username:     email,
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(input.DisplayName),
		Email:        email,
		Role:         "user",
		Status:       "active",
	}
	if user.DisplayName == "" {
		user.DisplayName = email
	}
	if err := service.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	service.logAudit(ctx, user.ID, "auth.register", user.Email, "", "")
	return user, nil
}

func (service *AuthService) ResolveSession(ctx context.Context, rawSessionToken string) (*model.User, *model.UserSession, error) {
	if strings.TrimSpace(rawSessionToken) == "" {
		return nil, nil, ErrInvalidSession
	}
	session, err := service.store.FindSessionByHash(ctx, security.HashToken(rawSessionToken))
	if err != nil {
		return nil, nil, err
	}
	if session == nil || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil, ErrInvalidSession
	}
	user, err := service.store.FindUserByID(ctx, session.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, nil, ErrInvalidSession
	}
	if time.Since(session.LastSeenAt) > 5*time.Minute {
		session.LastSeenAt = time.Now().UTC()
		_ = service.store.SaveSession(ctx, session)
	}
	return user, session, nil
}

func (service *AuthService) Logout(ctx context.Context, rawSessionToken string) error {
	session, err := service.store.FindSessionByHash(ctx, security.HashToken(rawSessionToken))
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}
	now := time.Now().UTC()
	session.RevokedAt = &now
	if err := service.store.SaveSession(ctx, session); err != nil {
		return err
	}
	service.logAudit(ctx, session.UserID, "auth.logout", session.ID, session.IPAddress, "")
	return nil
}

func (service *AuthService) logAudit(ctx context.Context, actorID string, action string, target string, ipAddress string, metadata string) {
	_ = service.store.CreateAuditLog(ctx, &model.AuditLog{
		ID:        security.NewID(),
		ActorID:   actorID,
		Action:    action,
		Target:    target,
		IPAddress: trimLength(ipAddress, 64),
		Metadata:  metadata,
	})
}

func (service *AuthService) findUserByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return nil, nil
	}
	if strings.Contains(trimmed, "@") {
		user, err := service.store.FindUserByEmail(ctx, strings.ToLower(trimmed))
		if err != nil {
			return nil, err
		}
		if user != nil {
			return user, nil
		}
	}
	return service.store.FindUserByUsername(ctx, trimmed)
}

func (service *AuthService) registrationEnabled(ctx context.Context) (bool, error) {
	setting, err := service.store.FindPlatformSetting(ctx, settingAllowRegistration)
	if err != nil {
		return false, err
	}
	if setting == nil {
		return false, nil
	}
	value := strings.TrimSpace(strings.ToLower(setting.Value))
	return value == "true" || value == "1", nil
}

func trimLength(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max]
}
