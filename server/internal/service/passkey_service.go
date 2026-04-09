package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/pkg/security"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
	"github.com/go-webauthn/webauthn/protocol"
	libwebauthn "github.com/go-webauthn/webauthn/webauthn"
)

const (
	passkeyCeremonyPurposeRegister = "register"
	passkeyCeremonyPurposeLogin    = "login"
	passkeyDefaultNamePrefix       = "通行密钥"
)

type PasskeyRecord struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

type PasskeyService struct {
	store       *repository.Store
	authService *AuthService
	webauthn    *libwebauthn.WebAuthn
}

type passkeyUser struct {
	user        *model.User
	credentials []libwebauthn.Credential
	handle      string
}

func NewPasskeyService(store *repository.Store, cfg config.Config, authService *AuthService) (*PasskeyService, error) {
	webauthnInstance, err := libwebauthn.New(&libwebauthn.Config{
		RPDisplayName: strings.TrimSpace(cfg.AppName),
		RPID:          strings.TrimSpace(cfg.WebAuthnRPID),
		RPOrigins:     cfg.WebAuthnRPOrigins,
	})
	if err != nil {
		return nil, fmt.Errorf("build passkey service: %w", err)
	}
	return &PasskeyService{store: store, authService: authService, webauthn: webauthnInstance}, nil
}

func (service *PasskeyService) BeginRegistration(ctx context.Context, actor *model.User) (string, *protocol.CredentialCreation, error) {
	passkeyUser, _, err := service.loadPasskeyUser(ctx, actor)
	if err != nil {
		return "", nil, err
	}
	creation, sessionData, err := service.webauthn.BeginMediatedRegistration(passkeyUser, protocol.MediationDefault,
		libwebauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyRequired(),
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}),
		libwebauthn.WithExclusions(libwebauthn.Credentials(passkeyUser.WebAuthnCredentials()).CredentialDescriptors()),
		libwebauthn.WithExtensions(map[string]any{"credProps": true}),
	)
	if err != nil {
		return "", nil, err
	}
	sessionID, err := service.createCeremony(ctx, passkeyCeremonyPurposeRegister, &actor.ID, sessionData)
	if err != nil {
		return "", nil, err
	}
	return sessionID, creation, nil
}

func (service *PasskeyService) FinishRegistration(ctx context.Context, actor *model.User, sessionID string, name string, credentialPayload []byte) (PasskeyRecord, error) {
	passkeyUser, existingRows, err := service.loadPasskeyUser(ctx, actor)
	if err != nil {
		return PasskeyRecord{}, err
	}
	sessionData, err := service.consumeCeremony(ctx, sessionID, passkeyCeremonyPurposeRegister, &actor.ID)
	if err != nil {
		return PasskeyRecord{}, err
	}
	request, err := httpRequestFromJSON(ctx, credentialPayload)
	if err != nil {
		return PasskeyRecord{}, err
	}
	credential, err := service.webauthn.FinishRegistration(passkeyUser, *sessionData, request)
	if err != nil {
		return PasskeyRecord{}, ErrInvalidGrant
	}
	credentialJSON, err := json.Marshal(credential)
	if err != nil {
		return PasskeyRecord{}, err
	}
	credentialID := encodeCredentialID(credential.ID)
	existing, err := service.store.FindUserPasskeyCredentialByCredentialID(ctx, credentialID)
	if err != nil {
		return PasskeyRecord{}, err
	}
	if existing != nil {
		return PasskeyRecord{}, ErrConflict
	}
	record := &model.UserPasskeyCredential{
		ID:             security.NewID(),
		UserID:         actor.ID,
		Name:           defaultPasskeyName(name, len(existingRows)+1),
		CredentialID:   credentialID,
		CredentialJSON: string(credentialJSON),
	}
	if err := service.store.CreateUserPasskeyCredential(ctx, record); err != nil {
		return PasskeyRecord{}, err
	}
	service.authService.logAudit(ctx, actor.ID, "auth.passkey.register", record.ID, "", "")
	return passkeyRecord(record), nil
}

func (service *PasskeyService) BeginLogin(ctx context.Context) (string, *protocol.CredentialAssertion, error) {
	initialized, err := service.authService.IsInitialized(ctx)
	if err != nil {
		return "", nil, err
	}
	if !initialized {
		return "", nil, fmt.Errorf("%w: 系统尚未初始化，请先创建管理员账号", ErrConflict)
	}
	assertion, sessionData, err := service.webauthn.BeginDiscoverableMediatedLogin(protocol.MediationDefault, libwebauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return "", nil, err
	}
	sessionID, err := service.createCeremony(ctx, passkeyCeremonyPurposeLogin, nil, sessionData)
	if err != nil {
		return "", nil, err
	}
	return sessionID, assertion, nil
}

func (service *PasskeyService) FinishLogin(ctx context.Context, sessionID string, credentialPayload []byte, meta SessionMeta) (*model.User, string, *model.UserSession, error) {
	sessionData, err := service.consumeCeremony(ctx, sessionID, passkeyCeremonyPurposeLogin, nil)
	if err != nil {
		return nil, "", nil, err
	}
	request, err := httpRequestFromJSON(ctx, credentialPayload)
	if err != nil {
		return nil, "", nil, err
	}
	validatedUser, validatedCredential, err := service.webauthn.FinishPasskeyLogin(func(rawID []byte, userHandle []byte) (libwebauthn.User, error) {
		if len(userHandle) == 0 {
			return nil, ErrUnauthorized
		}
		user, err := service.store.FindUserByWebAuthnUserHandle(ctx, string(userHandle))
		if err != nil {
			return nil, err
		}
		if user == nil || user.Status != "active" {
			return nil, ErrUnauthorized
		}
		wrappedUser, _, err := service.loadPasskeyUser(ctx, user)
		if err != nil {
			return nil, err
		}
		return wrappedUser, nil
	}, *sessionData, request)
	if err != nil {
		return nil, "", nil, ErrUnauthorized
	}
	userRecord, ok := validatedUser.(*passkeyUser)
	if !ok {
		return nil, "", nil, ErrUnauthorized
	}
	credentialRow, err := service.store.FindUserPasskeyCredentialByCredentialID(ctx, encodeCredentialID(validatedCredential.ID))
	if err != nil {
		return nil, "", nil, err
	}
	if credentialRow == nil || credentialRow.UserID != userRecord.user.ID {
		return nil, "", nil, ErrUnauthorized
	}
	credentialJSON, err := json.Marshal(validatedCredential)
	if err != nil {
		return nil, "", nil, err
	}
	now := time.Now().UTC()
	credentialRow.CredentialJSON = string(credentialJSON)
	credentialRow.LastUsedAt = &now
	if err := service.store.SaveUserPasskeyCredential(ctx, credentialRow); err != nil {
		return nil, "", nil, err
	}
	rawSessionToken, session, err := service.authService.createSession(ctx, userRecord.user, meta, "auth.passkey.login")
	if err != nil {
		return nil, "", nil, err
	}
	return userRecord.user, rawSessionToken, session, nil
}

func (service *PasskeyService) ListPasskeys(ctx context.Context, actor *model.User) ([]PasskeyRecord, error) {
	if actor == nil {
		return nil, ErrUnauthorized
	}
	credentials, err := service.store.ListUserPasskeyCredentials(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	items := make([]PasskeyRecord, 0, len(credentials))
	for _, credential := range credentials {
		items = append(items, passkeyRecord(&credential))
	}
	return items, nil
}

func (service *PasskeyService) DeletePasskey(ctx context.Context, actor *model.User, id string) error {
	if actor == nil {
		return ErrUnauthorized
	}
	credential, err := service.store.FindUserPasskeyCredentialByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if credential == nil {
		return ErrNotFound
	}
	if credential.UserID != actor.ID {
		return ErrForbidden
	}
	if err := service.store.DeleteUserPasskeyCredentialByID(ctx, credential.ID); err != nil {
		return err
	}
	service.authService.logAudit(ctx, actor.ID, "auth.passkey.delete", credential.ID, "", "")
	return nil
}

func (user *passkeyUser) WebAuthnID() []byte {
	return []byte(user.handle)
}

func (user *passkeyUser) WebAuthnName() string {
	return user.user.Username
}

func (user *passkeyUser) WebAuthnDisplayName() string {
	return user.user.Username
}

func (user *passkeyUser) WebAuthnCredentials() []libwebauthn.Credential {
	return user.credentials
}

func (service *PasskeyService) loadPasskeyUser(ctx context.Context, actor *model.User) (*passkeyUser, []model.UserPasskeyCredential, error) {
	if actor == nil {
		return nil, nil, ErrUnauthorized
	}
	user, err := service.store.FindUserByID(ctx, actor.ID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, nil, ErrUnauthorized
	}
	handle, err := service.ensureUserHandle(ctx, user)
	if err != nil {
		return nil, nil, err
	}
	credentialRows, err := service.store.ListUserPasskeyCredentials(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	credentials := make([]libwebauthn.Credential, 0, len(credentialRows))
	for _, row := range credentialRows {
		credential, err := decodeStoredCredential(row.CredentialJSON)
		if err != nil {
			return nil, nil, err
		}
		credentials = append(credentials, credential)
	}
	return &passkeyUser{user: user, credentials: credentials, handle: handle}, credentialRows, nil
}

func (service *PasskeyService) ensureUserHandle(ctx context.Context, user *model.User) (string, error) {
	if user == nil {
		return "", ErrUnauthorized
	}
	if user.WebAuthnUserHandle != nil {
		handle := strings.TrimSpace(*user.WebAuthnUserHandle)
		if handle != "" {
			if len(handle) > 64 {
				return "", fmt.Errorf("%w: 通行密钥用户标识无效", ErrInvalidInput)
			}
			return handle, nil
		}
	}
	handle, err := security.NewOpaqueToken(32)
	if err != nil {
		return "", err
	}
	user.WebAuthnUserHandle = &handle
	if err := service.store.SaveUser(ctx, user); err != nil {
		return "", err
	}
	return handle, nil
}

func (service *PasskeyService) createCeremony(ctx context.Context, purpose string, userID *uint, sessionData *libwebauthn.SessionData) (string, error) {
	encodedSession, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}
	id := security.NewID()
	expiresAt := sessionData.Expires
	if expiresAt.IsZero() {
		expiresAt = time.Now().UTC().Add(10 * time.Minute)
	}
	ceremony := &model.WebAuthnCeremony{
		ID:          id,
		Purpose:     purpose,
		UserID:      userID,
		SessionData: string(encodedSession),
		ExpiresAt:   expiresAt,
	}
	if err := service.store.CreateWebAuthnCeremony(ctx, ceremony); err != nil {
		return "", err
	}
	return id, nil
}

func (service *PasskeyService) consumeCeremony(ctx context.Context, sessionID string, purpose string, userID *uint) (*libwebauthn.SessionData, error) {
	ceremony, err := service.store.FindWebAuthnCeremonyByID(ctx, strings.TrimSpace(sessionID))
	if err != nil {
		return nil, err
	}
	if ceremony == nil || ceremony.Purpose != purpose || ceremony.ConsumedAt != nil || ceremony.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrInvalidGrant
	}
	if userID != nil {
		if ceremony.UserID == nil || *ceremony.UserID != *userID {
			return nil, ErrForbidden
		}
	}
	now := time.Now().UTC()
	ceremony.ConsumedAt = &now
	if err := service.store.SaveWebAuthnCeremony(ctx, ceremony); err != nil {
		return nil, err
	}
	var sessionData libwebauthn.SessionData
	if err := json.Unmarshal([]byte(ceremony.SessionData), &sessionData); err != nil {
		return nil, err
	}
	return &sessionData, nil
}

func httpRequestFromJSON(ctx context.Context, payload []byte) (*http.Request, error) {
	trimmed := bytes.TrimSpace(payload)
	if len(trimmed) == 0 {
		return nil, fmt.Errorf("%w: 缺少通行密钥响应数据", ErrInvalidInput)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", bytes.NewReader(trimmed))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	return request, nil
}

func decodeStoredCredential(raw string) (libwebauthn.Credential, error) {
	var credential libwebauthn.Credential
	if err := json.Unmarshal([]byte(raw), &credential); err != nil {
		return credential, err
	}
	return credential, nil
}

func encodeCredentialID(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func defaultPasskeyName(name string, ordinal int) string {
	trimmed := strings.TrimSpace(name)
	if trimmed != "" {
		return trimLength(trimmed, 160)
	}
	return fmt.Sprintf("%s %d", passkeyDefaultNamePrefix, ordinal)
}

func passkeyRecord(record *model.UserPasskeyCredential) PasskeyRecord {
	return PasskeyRecord{
		ID:         record.ID,
		Name:       record.Name,
		CreatedAt:  record.CreatedAt,
		UpdatedAt:  record.UpdatedAt,
		LastUsedAt: record.LastUsedAt,
	}
}
