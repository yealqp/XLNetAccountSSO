package service

import (
	"context"
	"strings"
	"time"

	"github.com/xianlin-network/sso-platform/server/internal/config"
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
	cfg   config.Config
}

func NewAuthService(store *repository.Store, cfg config.Config) *AuthService {
	return &AuthService{store: store, cfg: cfg}
}

func (service *AuthService) SeedDefaults(ctx context.Context) error {
	admin, err := service.store.FindUserByUsername(ctx, service.cfg.SeedAdminUsername)
	if err != nil {
		return err
	}
	if admin == nil {
		hash, err := security.HashPassword(service.cfg.SeedAdminPassword)
		if err != nil {
			return err
		}
		admin = &model.User{
			ID:           security.NewID(),
			Username:     service.cfg.SeedAdminUsername,
			PasswordHash: hash,
			DisplayName:  service.cfg.SeedAdminDisplayName,
			Email:        service.cfg.SeedAdminEmail,
			Role:         "admin",
			Status:       "active",
		}
		if err := service.store.CreateUser(ctx, admin); err != nil {
			return err
		}
	}

	demoClient, err := service.store.FindClientByClientID(ctx, service.cfg.SeedDemoClientID)
	if err != nil {
		return err
	}
	if demoClient == nil {
		demoClient = &model.OAuthClient{
			ID:          security.NewID(),
			Name:        service.cfg.SeedDemoClientName,
			Description: "Public SPA client used to verify the PKCE flow.",
			ClientID:    service.cfg.SeedDemoClientID,
			ClientType:  "public",
			Trusted:     false,
			CreatedBy:   admin.ID,
		}
		demoClient.SetRedirectURIs([]string{service.cfg.SeedDemoClientRedirect})
		demoClient.SetScopes([]string{"openid", "profile", "email", "offline_access"})
		if err := service.store.CreateClient(ctx, demoClient); err != nil {
			return err
		}
	} else if !contains(demoClient.Scopes(), "openid") {
		scopes, err := NormalizeKnownScopes(append([]string{"openid"}, demoClient.Scopes()...))
		if err != nil {
			return err
		}
		demoClient.SetScopes(scopes)
		if err := service.store.SaveClient(ctx, demoClient); err != nil {
			return err
		}
	}

	return nil
}

func (service *AuthService) Login(ctx context.Context, username string, password string, meta SessionMeta) (*model.User, string, *model.UserSession, error) {
	user, err := service.store.FindUserByUsername(ctx, strings.TrimSpace(username))
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

func trimLength(value string, max int) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= max {
		return trimmed
	}
	return trimmed[:max]
}
