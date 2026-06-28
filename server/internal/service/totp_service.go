package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/pkg/security"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
	"github.com/pquerna/otp/totp"
	"github.com/skip2/go-qrcode"
)

const totpIssuer = "XLNetAccount"

type TOTPService struct {
	store       *repository.Store
	cfg         config.Config
	authService *AuthService
}

type TOTPSetupResult struct {
	Secret    string `json:"secret"`
	URI       string `json:"uri"`
	QRDataURI string `json:"qr_data_uri"`
	SessionID string `json:"setup_session_id"`
}

type TOTPStatusResult struct {
	Enabled bool `json:"enabled"`
}

func NewTOTPService(store *repository.Store, cfg config.Config, authService *AuthService) *TOTPService {
	return &TOTPService{store: store, cfg: cfg, authService: authService}
}

func (service *TOTPService) BeginSetup(user *model.User) (*TOTPSetupResult, error) {
	if user == nil {
		return nil, ErrUnauthorized
	}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      totpIssuer,
		AccountName: user.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("generate totp key: %w", err)
	}
	qrBytes, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("encode qr code: %w", err)
	}
	qrDataURI := "data:image/png;base64," + base64.StdEncoding.EncodeToString(qrBytes)

	now := time.Now().UTC()
	ceremonyID := security.NewID()
	ceremony := &model.WebAuthnCeremony{
		ID:          ceremonyID,
		Purpose:     "totp_setup",
		UserID:      &user.ID,
		SessionData: key.Secret(),
		ExpiresAt:   now.Add(10 * time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.store.CreateWebAuthnCeremony(context.Background(), ceremony); err != nil {
		return nil, fmt.Errorf("create totp setup ceremony: %w", err)
	}

	return &TOTPSetupResult{
		Secret:    key.Secret(),
		URI:       key.URL(),
		QRDataURI: qrDataURI,
		SessionID: ceremonyID,
	}, nil
}

func (service *TOTPService) VerifySetup(actor *model.User, ceremonyID string, code string) error {
	if actor == nil {
		return ErrUnauthorized
	}
	ceremony, err := service.store.FindWebAuthnCeremonyByID(context.Background(), ceremonyID)
	if err != nil {
		return err
	}
	if ceremony == nil || ceremony.Purpose != "totp_setup" || ceremony.ConsumedAt != nil || ceremony.ExpiresAt.Before(time.Now().UTC()) {
		return ErrInvalidGrant
	}
	if ceremony.UserID == nil || *ceremony.UserID != actor.ID {
		return ErrForbidden
	}
	secret := ceremony.SessionData
	now := time.Now().UTC()
	ceremony.ConsumedAt = &now
	if err := service.store.SaveWebAuthnCeremony(context.Background(), ceremony); err != nil {
		return err
	}
	if !totp.Validate(code, secret) {
		return fmt.Errorf("%w: 验证码错误", ErrInvalidGrant)
	}
	existing, err := service.store.FindUserTOTPByUserID(context.Background(), actor.ID)
	if err != nil {
		return err
	}
	if existing != nil {
		existing.Secret = secret
		existing.Enabled = true
		existing.LastUsedAt = &now
		return service.store.SaveUserTOTP(context.Background(), existing)
	}
	totpRecord := &model.UserTOTP{
		ID:         security.NewID(),
		UserID:     actor.ID,
		Secret:     secret,
		Enabled:    true,
		LastUsedAt: &now,
	}
	return service.store.CreateUserTOTP(context.Background(), totpRecord)
}

func (service *TOTPService) Disable(actor *model.User) error {
	if actor == nil {
		return ErrUnauthorized
	}
	return service.store.DeleteUserTOTPByUserID(context.Background(), actor.ID)
}

func (service *TOTPService) Status(actor *model.User) (*TOTPStatusResult, error) {
	if actor == nil {
		return nil, ErrUnauthorized
	}
	totpRecord, err := service.store.FindUserTOTPByUserID(context.Background(), actor.ID)
	if err != nil {
		return nil, err
	}
	if totpRecord == nil || !totpRecord.Enabled {
		return &TOTPStatusResult{Enabled: false}, nil
	}
	return &TOTPStatusResult{Enabled: true}, nil
}

func (service *TOTPService) NeedsTOTP(user *model.User) (bool, error) {
	if user == nil {
		return false, ErrUnauthorized
	}
	totpRecord, err := service.store.FindUserTOTPByUserID(context.Background(), user.ID)
	if err != nil {
		return false, err
	}
	return totpRecord != nil && totpRecord.Enabled, nil
}

func (service *TOTPService) BeginTOTPLogin(user *model.User, meta SessionMeta) (string, error) {
	if user == nil {
		return "", ErrUnauthorized
	}
	now := time.Now().UTC()
	ceremonyID := security.NewID()
	ceremony := &model.WebAuthnCeremony{
		ID:          ceremonyID,
		Purpose:     "totp_login",
		UserID:      &user.ID,
		SessionData: fmt.Sprintf("%s|%s", meta.IPAddress, meta.UserAgent),
		ExpiresAt:   now.Add(5 * time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.store.CreateWebAuthnCeremony(context.Background(), ceremony); err != nil {
		return "", fmt.Errorf("create totp login ceremony: %w", err)
	}
	return ceremonyID, nil
}

func (service *TOTPService) VerifyTOTPLogin(ceremonyID string, code string, meta SessionMeta) (*model.User, string, *model.UserSession, error) {
	ceremony, err := service.store.FindWebAuthnCeremonyByID(context.Background(), ceremonyID)
	if err != nil {
		return nil, "", nil, err
	}
	if ceremony == nil || ceremony.Purpose != "totp_login" || ceremony.ConsumedAt != nil || ceremony.ExpiresAt.Before(time.Now().UTC()) {
		return nil, "", nil, ErrInvalidGrant
	}
	if ceremony.UserID == nil {
		return nil, "", nil, ErrInvalidGrant
	}
	user, err := service.store.FindUserByID(context.Background(), *ceremony.UserID)
	if err != nil {
		return nil, "", nil, err
	}
	if user == nil || user.Status != "active" {
		return nil, "", nil, ErrUnauthorized
	}
	totpRecord, err := service.store.FindUserTOTPByUserID(context.Background(), user.ID)
	if err != nil {
		return nil, "", nil, err
	}
	if totpRecord == nil || !totpRecord.Enabled {
		return nil, "", nil, ErrInvalidGrant
	}
	if !totp.Validate(code, totpRecord.Secret) {
		return nil, "", nil, fmt.Errorf("%w: 验证码错误", ErrInvalidGrant)
	}
	now := time.Now().UTC()
	ceremony.ConsumedAt = &now
	if err := service.store.SaveWebAuthnCeremony(context.Background(), ceremony); err != nil {
		return nil, "", nil, err
	}
	totpRecord.LastUsedAt = &now
	if err := service.store.SaveUserTOTP(context.Background(), totpRecord); err != nil {
		return nil, "", nil, err
	}

	rawSessionToken, session, err := service.authService.createSession(context.Background(), user, meta, "auth.login.totp")
	if err != nil {
		return nil, "", nil, err
	}
	return user, rawSessionToken, session, nil
}
