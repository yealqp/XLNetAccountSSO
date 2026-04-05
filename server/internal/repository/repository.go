package repository

import (
	"context"
	"errors"
	"time"

	"github.com/xianlin-network/sso-platform/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (store *Store) DB() *gorm.DB {
	return store.db
}

func (store *Store) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := store.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (store *Store) FindUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := store.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (store *Store) FindUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := store.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

func (store *Store) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := store.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (store *Store) ListUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := store.db.WithContext(ctx).Order("created_at desc").Find(&users).Error
	return users, err
}

func (store *Store) CreateUser(ctx context.Context, user *model.User) error {
	return store.db.WithContext(ctx).Create(user).Error
}

func (store *Store) SaveUser(ctx context.Context, user *model.User) error {
	return store.db.WithContext(ctx).Save(user).Error
}

func (store *Store) DeleteUser(ctx context.Context, id string) error {
	return store.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id).Error
}

func (store *Store) FindPlatformSetting(ctx context.Context, key string) (*model.PlatformSetting, error) {
	var setting model.PlatformSetting
	err := store.db.WithContext(ctx).Where("`key` = ?", key).First(&setting).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &setting, err
}

func (store *Store) SavePlatformSetting(ctx context.Context, setting *model.PlatformSetting) error {
	return store.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(setting).Error
}

func (store *Store) CreateEmailVerificationCode(ctx context.Context, code *model.EmailVerificationCode) error {
	return store.db.WithContext(ctx).Create(code).Error
}

func (store *Store) SaveEmailVerificationCode(ctx context.Context, code *model.EmailVerificationCode) error {
	return store.db.WithContext(ctx).Save(code).Error
}

func (store *Store) DeleteEmailVerificationCodes(ctx context.Context, email string, purpose string) error {
	return store.db.WithContext(ctx).Delete(&model.EmailVerificationCode{}, "email = ? AND purpose = ?", email, purpose).Error
}

func (store *Store) FindLatestEmailVerificationCode(ctx context.Context, email string, purpose string) (*model.EmailVerificationCode, error) {
	var code model.EmailVerificationCode
	err := store.db.WithContext(ctx).
		Where("email = ? AND purpose = ?", email, purpose).
		Order("created_at desc").
		First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &code, err
}

func (store *Store) FindClientByClientID(ctx context.Context, clientID string) (*model.OAuthClient, error) {
	var client model.OAuthClient
	err := store.db.WithContext(ctx).Where("client_id = ?", clientID).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &client, err
}

func (store *Store) FindClientByID(ctx context.Context, id string) (*model.OAuthClient, error) {
	var client model.OAuthClient
	err := store.db.WithContext(ctx).Where("id = ?", id).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &client, err
}

func (store *Store) ListClients(ctx context.Context) ([]model.OAuthClient, error) {
	var clients []model.OAuthClient
	err := store.db.WithContext(ctx).Order("created_at desc").Find(&clients).Error
	return clients, err
}

func (store *Store) CreateClient(ctx context.Context, client *model.OAuthClient) error {
	return store.db.WithContext(ctx).Create(client).Error
}

func (store *Store) SaveClient(ctx context.Context, client *model.OAuthClient) error {
	return store.db.WithContext(ctx).Save(client).Error
}

func (store *Store) DeleteClient(ctx context.Context, id string) error {
	return store.db.WithContext(ctx).Delete(&model.OAuthClient{}, "id = ?", id).Error
}

func (store *Store) CreateSession(ctx context.Context, session *model.UserSession) error {
	return store.db.WithContext(ctx).Create(session).Error
}

func (store *Store) FindSessionByHash(ctx context.Context, hash string) (*model.UserSession, error) {
	var session model.UserSession
	err := store.db.WithContext(ctx).Where("session_token_hash = ?", hash).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &session, err
}

func (store *Store) SaveSession(ctx context.Context, session *model.UserSession) error {
	return store.db.WithContext(ctx).Save(session).Error
}

func (store *Store) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	return store.db.WithContext(ctx).Delete(&model.UserSession{}, "user_id = ?", userID).Error
}

func (store *Store) CreateAuthorizationCode(ctx context.Context, code *model.AuthorizationCode) error {
	return store.db.WithContext(ctx).Create(code).Error
}

func (store *Store) FindAuthorizationCodeByHash(ctx context.Context, hash string) (*model.AuthorizationCode, error) {
	var code model.AuthorizationCode
	err := store.db.WithContext(ctx).Where("code_hash = ?", hash).First(&code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &code, err
}

func (store *Store) SaveAuthorizationCode(ctx context.Context, code *model.AuthorizationCode) error {
	return store.db.WithContext(ctx).Save(code).Error
}

func (store *Store) DeleteAuthorizationCodesByClientID(ctx context.Context, clientID string) error {
	return store.db.WithContext(ctx).Delete(&model.AuthorizationCode{}, "client_id = ?", clientID).Error
}

func (store *Store) DeleteAuthorizationCodesByUserID(ctx context.Context, userID string) error {
	return store.db.WithContext(ctx).Delete(&model.AuthorizationCode{}, "user_id = ?", userID).Error
}

func (store *Store) CreateAccessToken(ctx context.Context, token *model.AccessToken) error {
	return store.db.WithContext(ctx).Create(token).Error
}

func (store *Store) FindAccessTokenByHash(ctx context.Context, hash string) (*model.AccessToken, error) {
	var token model.AccessToken
	err := store.db.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

func (store *Store) FindAccessTokenByID(ctx context.Context, id string) (*model.AccessToken, error) {
	var token model.AccessToken
	err := store.db.WithContext(ctx).Where("id = ?", id).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

func (store *Store) ListAccessTokensByUser(ctx context.Context, userID string) ([]model.AccessToken, error) {
	var tokens []model.AccessToken
	err := store.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

func (store *Store) ListAccessTokensByUserAndClient(ctx context.Context, userID string, clientID string) ([]model.AccessToken, error) {
	var tokens []model.AccessToken
	err := store.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ?", userID, clientID).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

func (store *Store) SaveAccessToken(ctx context.Context, token *model.AccessToken) error {
	return store.db.WithContext(ctx).Save(token).Error
}

func (store *Store) DeleteAccessTokensByClientID(ctx context.Context, clientID string) error {
	return store.db.WithContext(ctx).Delete(&model.AccessToken{}, "client_id = ?", clientID).Error
}

func (store *Store) DeleteAccessTokensByUserID(ctx context.Context, userID string) error {
	return store.db.WithContext(ctx).Delete(&model.AccessToken{}, "user_id = ?", userID).Error
}

func (store *Store) CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	return store.db.WithContext(ctx).Create(token).Error
}

func (store *Store) FindRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := store.db.WithContext(ctx).Where("token_hash = ?", hash).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

func (store *Store) FindRefreshTokenByID(ctx context.Context, id string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	err := store.db.WithContext(ctx).Where("id = ?", id).First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &token, err
}

func (store *Store) ListRefreshTokensByUser(ctx context.Context, userID string) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := store.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

func (store *Store) ListRefreshTokensByUserAndClient(ctx context.Context, userID string, clientID string) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := store.db.WithContext(ctx).
		Where("user_id = ? AND client_id = ?", userID, clientID).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

func (store *Store) ListRefreshTokensByAccessTokenID(ctx context.Context, accessTokenID string) ([]model.RefreshToken, error) {
	var tokens []model.RefreshToken
	err := store.db.WithContext(ctx).
		Where("access_token_id = ?", accessTokenID).
		Order("created_at desc").
		Find(&tokens).Error
	return tokens, err
}

func (store *Store) SaveRefreshToken(ctx context.Context, token *model.RefreshToken) error {
	return store.db.WithContext(ctx).Save(token).Error
}

func (store *Store) DeleteRefreshTokensByClientID(ctx context.Context, clientID string) error {
	return store.db.WithContext(ctx).Delete(&model.RefreshToken{}, "client_id = ?", clientID).Error
}

func (store *Store) DeleteRefreshTokensByUserID(ctx context.Context, userID string) error {
	return store.db.WithContext(ctx).Delete(&model.RefreshToken{}, "user_id = ?", userID).Error
}

func (store *Store) CreateAuditLog(ctx context.Context, logEntry *model.AuditLog) error {
	return store.db.WithContext(ctx).Create(logEntry).Error
}

type Overview struct {
	Users          int64 `json:"users"`
	Clients        int64 `json:"clients"`
	ActiveSessions int64 `json:"active_sessions"`
	AccessTokens   int64 `json:"access_tokens"`
}

func (store *Store) Overview(ctx context.Context) (Overview, error) {
	now := time.Now().UTC()
	var result Overview
	if err := store.db.WithContext(ctx).Model(&model.User{}).Count(&result.Users).Error; err != nil {
		return result, err
	}
	if err := store.db.WithContext(ctx).Model(&model.OAuthClient{}).Count(&result.Clients).Error; err != nil {
		return result, err
	}
	if err := store.db.WithContext(ctx).Model(&model.UserSession{}).
		Where("revoked_at IS NULL AND expires_at > ?", now).
		Count(&result.ActiveSessions).Error; err != nil {
		return result, err
	}
	if err := store.db.WithContext(ctx).Model(&model.AccessToken{}).
		Where("revoked_at IS NULL AND expires_at > ?", now).
		Count(&result.AccessTokens).Error; err != nil {
		return result, err
	}
	return result, nil
}
