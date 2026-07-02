package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/pkg/security"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
)

type OAuthProvider string

const (
	OAuthMicrosoft OAuthProvider = "microsoft"
	OAuthGoogle    OAuthProvider = "google"
	OAuthGitHub    OAuthProvider = "github"
)

type OAuthLoginService struct {
	store *repository.Store
	cfg   config.Config
}

func NewOAuthLoginService(store *repository.Store, cfg config.Config) *OAuthLoginService {
	return &OAuthLoginService{store: store, cfg: cfg}
}

// ProviderConfig returns OAuth client credentials and endpoints for a provider.
func (s *OAuthLoginService) ProviderConfig(provider OAuthProvider) (clientID, clientSecret, authURL, tokenURL, userInfoURL, redirectURL string, ok bool) {
	base := strings.TrimRight(s.cfg.ServerBaseURL, "/")
	switch provider {
	case OAuthMicrosoft:
		tenant := s.cfg.OAuthMicrosoftTenant
		return s.cfg.OAuthMicrosoftClientID, s.cfg.OAuthMicrosoftClientSecret,
			fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", tenant),
			fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", tenant),
			"https://graph.microsoft.com/v1.0/me",
			base + "/api/auth/oauth/microsoft/callback",
			s.cfg.OAuthMicrosoftClientID != "" && s.cfg.OAuthMicrosoftClientSecret != ""
	case OAuthGoogle:
		return s.cfg.OAuthGoogleClientID, s.cfg.OAuthGoogleClientSecret,
			"https://accounts.google.com/o/oauth2/v2/auth",
			"https://oauth2.googleapis.com/token",
			"https://www.googleapis.com/oauth2/v2/userinfo",
			base + "/api/auth/oauth/google/callback",
			s.cfg.OAuthGoogleClientID != "" && s.cfg.OAuthGoogleClientSecret != ""
	case OAuthGitHub:
		return s.cfg.OAuthGitHubClientID, s.cfg.OAuthGitHubClientSecret,
			"https://github.com/login/oauth/authorize",
			"https://github.com/login/oauth/access_token",
			"https://api.github.com/user",
			base + "/api/auth/oauth/github/callback",
			s.cfg.OAuthGitHubClientID != "" && s.cfg.OAuthGitHubClientSecret != ""
	default:
		return "", "", "", "", "", "", false
	}
}

// AuthorizeURL builds the redirect URL to the provider's OAuth consent page.
func (s *OAuthLoginService) AuthorizeURL(provider OAuthProvider, state string) (string, error) {
	clientID, _, authURL, _, _, redirectURL, ok := s.ProviderConfig(provider)
	if !ok {
		return "", fmt.Errorf("%w: %s OAuth 未配置", ErrInvalidInput, provider)
	}
	var scope string
	switch provider {
	case OAuthMicrosoft:
		scope = "openid email profile User.Read"
	case OAuthGoogle:
		scope = "openid email profile"
	case OAuthGitHub:
		scope = "read:user user:email"
	}
	v := url.Values{}
	v.Set("client_id", clientID)
	v.Set("response_type", "code")
	v.Set("redirect_uri", redirectURL)
	v.Set("scope", scope)
	v.Set("state", state)
	return authURL + "?" + v.Encode(), nil
}

// ExchangeCode handles OAuth callback: exchange code → token → fetch user info.
func (s *OAuthLoginService) ExchangeCode(ctx context.Context, provider OAuthProvider, code string) (*OAuthUserInfo, error) {
	clientID, clientSecret, _, tokenURL, userInfoURL, redirectURL, ok := s.ProviderConfig(provider)
	if !ok {
		return nil, fmt.Errorf("%w: %s OAuth 未配置", ErrInvalidInput, provider)
	}
	token, err := exchangeOAuthToken(tokenURL, clientID, code, redirectURL, clientSecret)
	if err != nil {
		return nil, err
	}
	return fetchOAuthUserInfo(userInfoURL, token, provider)
}

// FindOrCreateUser looks up or creates a user linked to an OAuth provider account.
func (s *OAuthLoginService) FindOrCreateUser(ctx context.Context, provider OAuthProvider, info *OAuthUserInfo) (*model.User, string, *model.UserSession, error) {
	linked, err := s.store.FindOAuthLinkedAccount(ctx, string(provider), info.ID)
	if err != nil {
		return nil, "", nil, err
	}

	var user *model.User

	if linked != nil {
		user, err = s.store.FindUserByID(ctx, linked.UserID)
		if err != nil || user == nil {
			return nil, "", nil, fmt.Errorf("linked user not found")
		}
	} else {
		if info.Email != "" {
			user, _ = s.store.FindUserByEmail(ctx, info.Email)
		}
		if user == nil {
			username := info.Name
			if username == "" {
				username = fmt.Sprintf("%s_%s", provider, info.ID)
			}
			username = sanitizeUsername(username)
			pwHash, _ := security.HashPassword(security.NewID())
			user = &model.User{
				Username:     username,
				PasswordHash: pwHash,
				Email:        info.Email,
				Role:         "user",
				Status:       "active",
			}
			if err := s.store.CreateUser(ctx, user); err != nil {
				return nil, "", nil, err
			}
		}
		linked = &model.OAuthLinkedAccount{
			ID:             security.NewID(),
			UserID:         user.ID,
			Provider:       string(provider),
			ProviderUserID: info.ID,
			Email:          info.Email,
			Name:           info.Name,
		}
		if err := s.store.CreateOAuthLinkedAccount(ctx, linked); err != nil {
			return nil, "", nil, err
		}
	}

	rawToken := security.NewID()
	session := &model.UserSession{
		ID:               security.NewID(),
		SessionTokenHash: security.HashToken(rawToken),
		UserID:           user.ID,
		ExpiresAt:        time.Now().UTC().Add(sessionLifetime),
	}
	if err := s.store.CreateSession(ctx, session); err != nil {
		return nil, "", nil, err
	}

	return user, rawToken, session, nil
}

type OAuthUserInfo struct {
	ID    string
	Email string
	Name  string
}

func exchangeOAuthToken(tokenURL, clientID, code, redirectURL, clientSecret string) (string, error) {
	v := url.Values{}
	v.Set("client_id", clientID)
	v.Set("code", code)
	v.Set("redirect_uri", redirectURL)
	v.Set("grant_type", "authorization_code")
	v.Set("client_secret", clientSecret)

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(v.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: token 交换请求失败", ErrInvalidInput)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("%w: token 响应解析失败", ErrInvalidInput)
	}
	if result.Error != "" {
		return "", fmt.Errorf("%w: token 交换失败（%s）", ErrInvalidInput, result.Error)
	}
	return result.AccessToken, nil
}

func fetchOAuthUserInfo(userInfoURL, accessToken string, provider OAuthProvider) (*OAuthUserInfo, error) {
	req, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if provider == OAuthGitHub {
		req.Header.Set("Accept", "application/vnd.github.v3+json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: 用户信息请求失败", ErrInvalidInput)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch provider {
	case OAuthMicrosoft:
		var info struct {
			ID                string `json:"id"`
			Mail              string `json:"mail"`
			UserPrincipalName string `json:"userPrincipalName"`
			DisplayName       string `json:"displayName"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, fmt.Errorf("%w: Microsoft 用户信息解析失败", ErrInvalidInput)
		}
		email := info.Mail
		if email == "" {
			email = info.UserPrincipalName
		}
		return &OAuthUserInfo{ID: info.ID, Email: email, Name: info.DisplayName}, nil
	case OAuthGoogle:
		var info struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, fmt.Errorf("%w: Google 用户信息解析失败", ErrInvalidInput)
		}
		return &OAuthUserInfo{ID: info.ID, Email: info.Email, Name: info.Name}, nil
	case OAuthGitHub:
		var info struct {
			ID    int    `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Login string `json:"login"`
		}
		if err := json.Unmarshal(body, &info); err != nil {
			return nil, fmt.Errorf("%w: GitHub 用户信息解析失败", ErrInvalidInput)
		}
		name := info.Name
		if name == "" {
			name = info.Login
		}
		return &OAuthUserInfo{ID: fmt.Sprintf("%d", info.ID), Email: info.Email, Name: name}, nil
	}
	return nil, fmt.Errorf("%w: unknown provider", ErrInvalidInput)
}

func sanitizeUsername(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	s := strings.Trim(b.String(), "_")
	if len(s) > 64 {
		s = s[:64]
	}
	if s == "" {
		s = "user"
	}
	return s
}
