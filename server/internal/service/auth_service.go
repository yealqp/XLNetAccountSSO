package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

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
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type RegisterInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

type UpdateProfileInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
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
	password := input.Password
	if username == "" || strings.TrimSpace(password) == "" {
		return nil, ErrInvalidInput
	}
	if err := validateRegistrationPassword(password); err != nil {
		return nil, err
	}
	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:     username,
		PasswordHash: hash,
		Email:        strings.TrimSpace(strings.ToLower(input.Email)),
		Role:         "admin",
		Status:       "active",
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
	username := strings.TrimSpace(input.Username)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := input.Password
	if username == "" {
		return nil, fmt.Errorf("%w: 请输入用户名", ErrInvalidInput)
	}
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: 请输入有效邮箱", ErrInvalidInput)
	}
	if strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("%w: 请输入密码", ErrInvalidInput)
	}
	if strings.TrimSpace(input.Code) == "" {
		return nil, fmt.Errorf("%w: 请输入验证码", ErrInvalidInput)
	}
	if err := validateRegistrationPassword(password); err != nil {
		return nil, err
	}
	existingByUsername, err := service.store.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingByUsername != nil {
		return nil, ErrConflict
	}
	existingByEmail, err := service.store.FindUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if existingByEmail != nil {
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
		Username:     username,
		PasswordHash: hash,
		Email:        email,
		Role:         "user",
		Status:       "active",
	}
	if err := service.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	service.logAudit(ctx, user.ID, "auth.register", user.Email, "", "")
	return user, nil
}

func (service *AuthService) UpdateProfile(ctx context.Context, actor *model.User, input UpdateProfileInput) (*model.User, error) {
	if actor == nil {
		return nil, ErrUnauthorized
	}
	user, err := service.store.FindUserByID(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	username := strings.TrimSpace(input.Username)
	if username == "" {
		return nil, fmt.Errorf("%w: 用户名不能为空", ErrInvalidInput)
	}
	if username != user.Username {
		existing, err := service.store.FindUserByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != user.ID {
			return nil, ErrConflict
		}
		user.Username = username
	}
	if strings.TrimSpace(input.Password) != "" {
		if strings.TrimSpace(input.Code) == "" {
			return nil, fmt.Errorf("%w: 修改密码需要邮箱验证码", ErrInvalidInput)
		}
		verificationService := NewVerificationService(service.store)
		if err := verificationService.VerifyProfilePasswordCode(ctx, user.Email, input.Code); err != nil {
			return nil, err
		}
		if err := validateRegistrationPassword(input.Password); err != nil {
			return nil, err
		}
		hash, err := security.HashPassword(input.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	if err := service.store.SaveUser(ctx, user); err != nil {
		return nil, err
	}
	service.logAudit(ctx, user.ID, "auth.profile.update", user.Username, "", "")
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

func (service *AuthService) logAudit(ctx context.Context, actorID uint, action string, target string, ipAddress string, metadata string) {
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

func validateRegistrationPassword(password string) error {
	if len(password) < 8 || len(password) > 20 {
		return fmt.Errorf("%w: 注册密码需为8-20位，且包含大小写字母和数字", ErrInvalidInput)
	}
	var hasUpper bool
	var hasLower bool
	var hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return fmt.Errorf("%w: 注册密码需为8-20位，且包含大小写字母和数字", ErrInvalidInput)
	}
	return nil
}
