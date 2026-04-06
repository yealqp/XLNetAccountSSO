package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/xianlin-network/sso-platform/server/internal/model"
	"github.com/xianlin-network/sso-platform/server/internal/pkg/security"
	"github.com/xianlin-network/sso-platform/server/internal/repository"
)

type TokenService struct {
	store *repository.Store
}

func NewTokenService(store *repository.Store) *TokenService {
	return &TokenService{store: store}
}

func (service *TokenService) ListTokens(ctx context.Context, actor *model.User) ([]map[string]any, error) {
	if actor == nil {
		return nil, ErrUnauthorized
	}
	accessTokens, err := service.store.ListAccessTokensByUser(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	refreshTokens, err := service.store.ListRefreshTokensByUser(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	return service.buildTokenItems(ctx, accessTokens, refreshTokens)
}

func (service *TokenService) ListAllTokens(ctx context.Context, actor *model.User) ([]map[string]any, error) {
	if actor == nil || actor.Role != "admin" {
		return nil, ErrForbidden
	}
	accessTokens, err := service.store.ListAccessTokens(ctx)
	if err != nil {
		return nil, err
	}
	refreshTokens, err := service.store.ListRefreshTokens(ctx)
	if err != nil {
		return nil, err
	}
	return service.buildTokenItems(ctx, accessTokens, refreshTokens)
}

func (service *TokenService) buildTokenItems(ctx context.Context, accessTokens []model.AccessToken, refreshTokens []model.RefreshToken) ([]map[string]any, error) {
	clientNames, err := service.clientNames(ctx, accessTokens, refreshTokens)
	if err != nil {
		return nil, err
	}
	ownerNames, err := service.userNames(ctx, accessTokens, refreshTokens)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	items := make([]map[string]any, 0, len(accessTokens)+len(refreshTokens))
	for _, token := range accessTokens {
		items = append(items, map[string]any{
			"id":             token.ID,
			"token_kind":     "access",
			"client_id":      token.ClientID,
			"client_name":    clientNames[token.ClientID],
			"user_id":        token.UserID,
			"owner_username": ownerNames[token.UserID],
			"scope":          token.Scope,
			"status":         tokenStatus(token.RevokedAt, token.ExpiresAt, now),
			"expires_at":     token.ExpiresAt,
			"revoked_at":     token.RevokedAt,
			"created_at":     token.CreatedAt,
			"updated_at":     token.UpdatedAt,
		})
	}
	for _, token := range refreshTokens {
		items = append(items, map[string]any{
			"id":                      token.ID,
			"token_kind":              "refresh",
			"client_id":               token.ClientID,
			"client_name":             clientNames[token.ClientID],
			"user_id":                 token.UserID,
			"owner_username":          ownerNames[token.UserID],
			"scope":                   token.Scope,
			"status":                  tokenStatus(token.RevokedAt, token.ExpiresAt, now),
			"expires_at":              token.ExpiresAt,
			"revoked_at":              token.RevokedAt,
			"created_at":              token.CreatedAt,
			"updated_at":              token.UpdatedAt,
			"related_access_token_id": token.AccessTokenID,
		})
	}

	slices.SortFunc(items, func(left map[string]any, right map[string]any) int {
		leftTime, _ := left["created_at"].(time.Time)
		rightTime, _ := right["created_at"].(time.Time)
		if leftTime.Equal(rightTime) {
			leftKind, _ := left["token_kind"].(string)
			rightKind, _ := right["token_kind"].(string)
			return compareStrings(leftKind, rightKind)
		}
		if leftTime.After(rightTime) {
			return -1
		}
		return 1
	})

	return items, nil
}

func (service *TokenService) RevokeAccessToken(ctx context.Context, actor *model.User, tokenID string) error {
	accessToken, err := service.store.FindAccessTokenByID(ctx, tokenID)
	if err != nil {
		return err
	}
	if accessToken == nil {
		return ErrNotFound
	}
	if !canManageToken(actor, accessToken.UserID) {
		return ErrForbidden
	}
	now := time.Now().UTC()
	if accessToken.RevokedAt == nil {
		accessToken.RevokedAt = &now
		if err := service.store.SaveAccessToken(ctx, accessToken); err != nil {
			return err
		}
	}
	refreshTokens, err := service.store.ListRefreshTokensByAccessTokenID(ctx, accessToken.ID)
	if err != nil {
		return err
	}
	for index := range refreshTokens {
		refreshToken := refreshTokens[index]
		if refreshToken.RevokedAt != nil {
			continue
		}
		refreshToken.RevokedAt = &now
		if err := service.store.SaveRefreshToken(ctx, &refreshToken); err != nil {
			return err
		}
	}
	service.logAudit(ctx, actor.ID, "token.revoke.access", accessToken.ID, "")
	return nil
}

func (service *TokenService) RevokeRefreshToken(ctx context.Context, actor *model.User, tokenID string) error {
	refreshToken, err := service.store.FindRefreshTokenByID(ctx, tokenID)
	if err != nil {
		return err
	}
	if refreshToken == nil {
		return ErrNotFound
	}
	if !canManageToken(actor, refreshToken.UserID) {
		return ErrForbidden
	}
	if refreshToken.RevokedAt != nil {
		return nil
	}
	now := time.Now().UTC()
	refreshToken.RevokedAt = &now
	if err := service.store.SaveRefreshToken(ctx, refreshToken); err != nil {
		return err
	}
	service.logAudit(ctx, actor.ID, "token.revoke.refresh", refreshToken.ID, "")
	return nil
}

func (service *TokenService) RevokeClientTokens(ctx context.Context, actor *model.User, clientID string) error {
	var (
		accessTokens  []model.AccessToken
		refreshTokens []model.RefreshToken
		err           error
	)
	if actor == nil {
		return ErrUnauthorized
	}
	if actor.Role == "admin" {
		accessTokens, err = service.store.ListAccessTokensByClientID(ctx, clientID)
	} else {
		accessTokens, err = service.store.ListAccessTokensByUserAndClient(ctx, actor.ID, clientID)
	}
	if err != nil {
		return err
	}
	if actor.Role == "admin" {
		refreshTokens, err = service.store.ListRefreshTokensByClientID(ctx, clientID)
	} else {
		refreshTokens, err = service.store.ListRefreshTokensByUserAndClient(ctx, actor.ID, clientID)
	}
	if err != nil {
		return err
	}
	if len(accessTokens) == 0 && len(refreshTokens) == 0 {
		return ErrNotFound
	}
	now := time.Now().UTC()
	for index := range accessTokens {
		accessToken := accessTokens[index]
		if accessToken.RevokedAt != nil {
			continue
		}
		accessToken.RevokedAt = &now
		if err := service.store.SaveAccessToken(ctx, &accessToken); err != nil {
			return err
		}
	}
	for index := range refreshTokens {
		refreshToken := refreshTokens[index]
		if refreshToken.RevokedAt != nil {
			continue
		}
		refreshToken.RevokedAt = &now
		if err := service.store.SaveRefreshToken(ctx, &refreshToken); err != nil {
			return err
		}
	}
	service.logAudit(ctx, actor.ID, "token.revoke.client", clientID, "")
	return nil
}

func (service *TokenService) clientNames(ctx context.Context, accessTokens []model.AccessToken, refreshTokens []model.RefreshToken) (map[string]string, error) {
	clientNames := map[string]string{}
	clientIDs := map[string]struct{}{}
	for _, token := range accessTokens {
		clientIDs[token.ClientID] = struct{}{}
	}
	for _, token := range refreshTokens {
		clientIDs[token.ClientID] = struct{}{}
	}
	for clientID := range clientIDs {
		client, err := service.store.FindClientByClientID(ctx, clientID)
		if err != nil {
			return nil, err
		}
		if client == nil {
			clientNames[clientID] = clientID
			continue
		}
		clientNames[clientID] = fmt.Sprintf("%s (%s)", client.Name, client.ClientType)
	}
	return clientNames, nil
}

func (service *TokenService) userNames(ctx context.Context, accessTokens []model.AccessToken, refreshTokens []model.RefreshToken) (map[uint]string, error) {
	userNames := map[uint]string{}
	userIDs := map[uint]struct{}{}
	for _, token := range accessTokens {
		userIDs[token.UserID] = struct{}{}
	}
	for _, token := range refreshTokens {
		userIDs[token.UserID] = struct{}{}
	}
	for userID := range userIDs {
		user, err := service.store.FindUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}
		if user == nil {
			userNames[userID] = ""
			continue
		}
		userNames[userID] = user.Username
	}
	return userNames, nil
}

func (service *TokenService) logAudit(ctx context.Context, actorID uint, action string, target string, metadata string) {
	_ = service.store.CreateAuditLog(ctx, &model.AuditLog{
		ID:       security.NewID(),
		ActorID:  actorID,
		Action:   action,
		Target:   target,
		Metadata: metadata,
	})
}

func canManageToken(actor *model.User, tokenOwnerID uint) bool {
	if actor == nil {
		return false
	}
	return actor.Role == "admin" || actor.ID == tokenOwnerID
}

func tokenStatus(revokedAt *time.Time, expiresAt time.Time, now time.Time) string {
	if revokedAt != nil {
		return "revoked"
	}
	if expiresAt.Before(now) {
		return "expired"
	}
	return "active"
}

func compareStrings(left string, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
