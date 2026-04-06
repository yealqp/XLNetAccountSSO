package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/pkg/security"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
)

const (
	authorizationCodeLifetime = 5 * time.Minute
	accessTokenLifetime       = 15 * time.Minute
	refreshTokenLifetime      = 14 * 24 * time.Hour
)

type AuthorizationRequest struct {
	ResponseType        string `json:"response_type"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	State               string `json:"state"`
	Nonce               string `json:"nonce"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
}

type AuthorizationDecision struct {
	AuthorizationRequest
	Approved bool `json:"approved"`
}

type TokenRequest struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	ClientSecret string
	CodeVerifier string
	RefreshToken string
}

type OAuthService struct {
	store  *repository.Store
	cfg    config.Config
	signer *security.OIDCSigner
}

func NewOAuthService(store *repository.Store, cfg config.Config) (*OAuthService, error) {
	signer, err := security.NewOIDCSigner(cfg.OIDCPrivateKeyPEM, cfg.OIDCKeyID)
	if err != nil {
		return nil, err
	}
	return &OAuthService{store: store, cfg: cfg, signer: signer}, nil
}

func (service *OAuthService) PreviewAuthorization(ctx context.Context, user *model.User, input AuthorizationRequest) (map[string]any, error) {
	if user == nil {
		return nil, ErrUnauthorized
	}
	client, scopeItems, err := service.validateAuthorizationRequest(ctx, input)
	if err != nil {
		return nil, err
	}
	if resolvedIconURL, syncErr := SyncClientIconURL(ctx, service.cfg, client.ID, client.IconURL); syncErr == nil && resolvedIconURL != client.IconURL {
		client.IconURL = resolvedIconURL
		_ = service.store.SaveClient(ctx, client)
	}
	return map[string]any{
		"client": map[string]any{
			"id":                      client.ID,
			"name":                    client.Name,
			"description":             client.Description,
			"icon_url":                client.IconURL,
			"client_id":               client.ClientID,
			"redirect_uri":            input.RedirectURI,
			"client_type":             client.ClientType,
			"trusted":                 client.Trusted,
			"requested_scopes":        scopeItems,
			"requested_scope_details": ScopeDetails(scopeItems),
		},
		"user": map[string]any{
			"id":       user.ID,
			"username": user.Username,
		},
		"request": map[string]any{
			"response_type":         input.ResponseType,
			"client_id":             input.ClientID,
			"redirect_uri":          input.RedirectURI,
			"scope":                 strings.Join(scopeItems, " "),
			"state":                 input.State,
			"nonce":                 input.Nonce,
			"code_challenge":        input.CodeChallenge,
			"code_challenge_method": input.CodeChallengeMethod,
		},
	}, nil
}

func (service *OAuthService) Authorize(ctx context.Context, user *model.User, decision AuthorizationDecision) (string, error) {
	if user == nil {
		return "", ErrUnauthorized
	}
	_, scopeItems, err := service.validateAuthorizationRequest(ctx, decision.AuthorizationRequest)
	if err != nil {
		return "", err
	}
	if !decision.Approved {
		return buildRedirect(decision.RedirectURI, map[string]string{
			"error":             "access_denied",
			"error_description": "The user denied the request.",
			"state":             decision.State,
		}), nil
	}
	rawCode, err := security.NewOpaqueToken(24)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	code := &model.AuthorizationCode{
		ID:                  security.NewID(),
		CodeHash:            security.HashToken(rawCode),
		ClientID:            decision.ClientID,
		UserID:              user.ID,
		RedirectURI:         decision.RedirectURI,
		Scope:               strings.Join(scopeItems, " "),
		State:               decision.State,
		Nonce:               strings.TrimSpace(decision.Nonce),
		CodeChallenge:       decision.CodeChallenge,
		CodeChallengeMethod: decision.CodeChallengeMethod,
		ExpiresAt:           now.Add(authorizationCodeLifetime),
	}
	if err := service.store.CreateAuthorizationCode(ctx, code); err != nil {
		return "", err
	}
	return buildRedirect(decision.RedirectURI, map[string]string{
		"code":  rawCode,
		"state": decision.State,
	}), nil
}

func (service *OAuthService) ExchangeToken(ctx context.Context, input TokenRequest) (map[string]any, error) {
	grantType := strings.TrimSpace(input.GrantType)
	switch grantType {
	case "authorization_code":
		return service.exchangeAuthorizationCode(ctx, input)
	case "refresh_token":
		return service.exchangeRefreshToken(ctx, input)
	default:
		return nil, ErrInvalidGrant
	}
}

func (service *OAuthService) UserInfo(ctx context.Context, rawToken string) (map[string]any, error) {
	accessToken, user, err := service.resolveAccessToken(ctx, rawToken)
	if err != nil {
		return nil, err
	}
	response := map[string]any{
		"sub": user.ID,
	}
	if HasScope(accessToken.Scope, "profile") {
		response["preferred_username"] = user.Username
		response["name"] = user.Username
	}
	if HasScope(accessToken.Scope, "email") {
		response["email"] = user.Email
	}
	if HasScope(accessToken.Scope, "roles") {
		response["roles"] = []string{user.Role}
	}
	return response, nil
}

func (service *OAuthService) RevokeToken(ctx context.Context, token string, clientID string) error {
	hash := security.HashToken(token)
	accessToken, err := service.store.FindAccessTokenByHash(ctx, hash)
	if err != nil {
		return err
	}
	if accessToken != nil {
		if clientID != "" && accessToken.ClientID != clientID {
			return ErrInvalidClient
		}
		now := time.Now().UTC()
		accessToken.RevokedAt = &now
		return service.store.SaveAccessToken(ctx, accessToken)
	}
	refreshToken, err := service.store.FindRefreshTokenByHash(ctx, hash)
	if err != nil {
		return err
	}
	if refreshToken == nil {
		return nil
	}
	if clientID != "" && refreshToken.ClientID != clientID {
		return ErrInvalidClient
	}
	now := time.Now().UTC()
	refreshToken.RevokedAt = &now
	return service.store.SaveRefreshToken(ctx, refreshToken)
}

func (service *OAuthService) Introspect(ctx context.Context, rawToken string) (map[string]any, error) {
	accessToken, user, err := service.resolveAccessToken(ctx, rawToken)
	if err == nil {
		return map[string]any{
			"active":    true,
			"sub":       user.ID,
			"username":  user.Username,
			"client_id": accessToken.ClientID,
			"scope":     accessToken.Scope,
			"exp":       accessToken.ExpiresAt.Unix(),
		}, nil
	}
	return map[string]any{"active": false}, nil
}

func (service *OAuthService) DiscoveryMetadata() map[string]any {
	baseURL := service.cfg.BaseURL()
	issuerURL := service.cfg.IssuerURL()
	return map[string]any{
		"claims_supported":                              []string{"sub", "iss", "aud", "exp", "iat", "nonce", "preferred_username", "name", "email", "roles"},
		"claim_types_supported":                         []string{"normal"},
		"code_challenge_methods_supported":              []string{"S256", "plain"},
		"grant_types_supported":                         []string{"authorization_code", "refresh_token"},
		"id_token_signing_alg_values_supported":         []string{"RS256"},
		"introspection_endpoint":                        baseURL + "/oauth/introspect",
		"introspection_endpoint_auth_methods_supported": []string{"none", "client_secret_basic", "client_secret_post"},
		"issuer":                   issuerURL,
		"jwks_uri":                 baseURL + "/.well-known/jwks.json",
		"response_modes_supported": []string{"query"},
		"response_types_supported": []string{"code"},
		"revocation_endpoint":      baseURL + "/oauth/revoke",
		"revocation_endpoint_auth_methods_supported": []string{"none", "client_secret_basic", "client_secret_post"},
		"scopes_supported":                           scopeKeys(ScopeDefinitions()),
		"subject_types_supported":                    []string{"public"},
		"token_endpoint":                             baseURL + "/oauth/token",
		"token_endpoint_auth_methods_supported":      []string{"none", "client_secret_basic", "client_secret_post"},
		"userinfo_endpoint":                          baseURL + "/oauth/userinfo",
		"authorization_endpoint":                     baseURL + "/oauth/authorize",
	}
}

func (service *OAuthService) JSONWebKeySet() map[string]any {
	return service.signer.JWKS()
}

func (service *OAuthService) exchangeAuthorizationCode(ctx context.Context, input TokenRequest) (map[string]any, error) {
	client, err := service.validateClient(ctx, input.ClientID, input.ClientSecret)
	if err != nil {
		return nil, err
	}
	code, err := service.store.FindAuthorizationCodeByHash(ctx, security.HashToken(strings.TrimSpace(input.Code)))
	if err != nil {
		return nil, err
	}
	if code == nil || code.ConsumedAt != nil || code.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrInvalidGrant
	}
	if code.ClientID != client.ClientID || code.RedirectURI != strings.TrimSpace(input.RedirectURI) {
		return nil, fmt.Errorf("%w: 授权码与客户端或回调地址不匹配", ErrInvalidGrant)
	}
	if code.CodeChallenge != "" || strings.TrimSpace(code.CodeChallengeMethod) != "" {
		codeVerifier := strings.TrimSpace(input.CodeVerifier)
		if codeVerifier == "" {
			return nil, fmt.Errorf("%w: 缺少 code_verifier", ErrInvalidGrant)
		}
		switch pkceMethod(code.CodeChallengeMethod) {
		case "S256":
			if security.S256CodeChallenge(codeVerifier) != code.CodeChallenge {
				return nil, fmt.Errorf("%w: code_verifier 校验失败", ErrInvalidGrant)
			}
		case "plain":
			if codeVerifier != code.CodeChallenge {
				return nil, fmt.Errorf("%w: code_verifier 校验失败", ErrInvalidGrant)
			}
		default:
			return nil, fmt.Errorf("%w: 不支持的 PKCE 校验方式", ErrInvalidGrant)
		}
	}
	response, err := service.issueTokens(ctx, client.ClientID, code.UserID, code.Scope, code.Nonce, true)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	code.ConsumedAt = &now
	if err := service.store.SaveAuthorizationCode(ctx, code); err != nil {
		return nil, err
	}
	return response, nil
}

func (service *OAuthService) exchangeRefreshToken(ctx context.Context, input TokenRequest) (map[string]any, error) {
	client, err := service.validateClient(ctx, input.ClientID, input.ClientSecret)
	if err != nil {
		return nil, err
	}
	refreshToken, err := service.store.FindRefreshTokenByHash(ctx, security.HashToken(strings.TrimSpace(input.RefreshToken)))
	if err != nil {
		return nil, err
	}
	if refreshToken == nil || refreshToken.RevokedAt != nil || refreshToken.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrInvalidGrant
	}
	if refreshToken.ClientID != client.ClientID {
		return nil, ErrInvalidGrant
	}
	now := time.Now().UTC()
	refreshToken.RevokedAt = &now
	if err := service.store.SaveRefreshToken(ctx, refreshToken); err != nil {
		return nil, err
	}
	if strings.TrimSpace(refreshToken.AccessTokenID) != "" {
		accessToken, err := service.store.FindAccessTokenByID(ctx, refreshToken.AccessTokenID)
		if err != nil {
			return nil, err
		}
		if accessToken != nil && accessToken.RevokedAt == nil {
			accessToken.RevokedAt = &now
			if err := service.store.SaveAccessToken(ctx, accessToken); err != nil {
				return nil, err
			}
		}
	}
	return service.issueTokens(ctx, refreshToken.ClientID, refreshToken.UserID, refreshToken.Scope, "", false)
}

func (service *OAuthService) issueTokens(ctx context.Context, clientID string, userID uint, scope string, nonce string, includeIDToken bool) (map[string]any, error) {
	normalizedScopes, err := NormalizeKnownScopes(normalizeScopeList(strings.Fields(strings.ReplaceAll(strings.TrimSpace(scope), ",", " "))))
	if err != nil {
		return nil, err
	}
	grantedScope := strings.Join(normalizedScopes, " ")
	rawAccessToken, err := security.NewOpaqueToken(32)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	accessToken := &model.AccessToken{
		ID:        security.NewID(),
		TokenHash: security.HashToken(rawAccessToken),
		ClientID:  clientID,
		UserID:    userID,
		Scope:     grantedScope,
		ExpiresAt: now.Add(accessTokenLifetime),
	}
	if err := service.store.CreateAccessToken(ctx, accessToken); err != nil {
		return nil, err
	}
	response := map[string]any{
		"access_token": rawAccessToken,
		"token_type":   "Bearer",
		"expires_in":   int(accessTokenLifetime.Seconds()),
		"scope":        grantedScope,
	}
	if hasScopeList(normalizedScopes, "offline_access") {
		rawRefreshToken, err := security.NewOpaqueToken(40)
		if err != nil {
			return nil, err
		}
		refreshToken := &model.RefreshToken{
			ID:            security.NewID(),
			TokenHash:     security.HashToken(rawRefreshToken),
			AccessTokenID: accessToken.ID,
			ClientID:      clientID,
			UserID:        userID,
			Scope:         grantedScope,
			ExpiresAt:     now.Add(refreshTokenLifetime),
		}
		if err := service.store.CreateRefreshToken(ctx, refreshToken); err != nil {
			return nil, err
		}
		response["refresh_token"] = rawRefreshToken
	}
	if includeIDToken && hasScopeList(normalizedScopes, "openid") {
		rawIDToken, err := service.issueIDToken(ctx, clientID, userID, grantedScope, nonce, now)
		if err != nil {
			return nil, err
		}
		response["id_token"] = rawIDToken
	}
	return response, nil
}

func (service *OAuthService) issueIDToken(ctx context.Context, clientID string, userID uint, scope string, nonce string, issuedAt time.Time) (string, error) {
	user, err := service.store.FindUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidGrant
	}
	claims := map[string]any{
		"aud": clientID,
		"exp": issuedAt.Add(accessTokenLifetime).Unix(),
		"iat": issuedAt.Unix(),
		"iss": service.cfg.IssuerURL(),
		"sub": user.ID,
	}
	if strings.TrimSpace(nonce) != "" {
		claims["nonce"] = strings.TrimSpace(nonce)
	}
	if HasScope(scope, "profile") {
		claims["preferred_username"] = user.Username
		claims["name"] = user.Username
	}
	if HasScope(scope, "email") {
		claims["email"] = user.Email
	}
	if HasScope(scope, "roles") {
		claims["roles"] = []string{user.Role}
	}
	return service.signer.SignJWT(claims)
}

func (service *OAuthService) validateClient(ctx context.Context, clientID string, clientSecret string) (*model.OAuthClient, error) {
	client, err := service.store.FindClientByClientID(ctx, strings.TrimSpace(clientID))
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrInvalidClient
	}
	if client.ClientType == "confidential" {
		if strings.TrimSpace(clientSecret) == "" || security.ComparePassword(client.ClientSecretHash, clientSecret) != nil {
			return nil, ErrInvalidClient
		}
	}
	return client, nil
}

func (service *OAuthService) validateAuthorizationRequest(ctx context.Context, input AuthorizationRequest) (*model.OAuthClient, []string, error) {
	if strings.TrimSpace(input.ResponseType) != "code" {
		return nil, nil, fmt.Errorf("%w: 仅支持 response_type=code", ErrInvalidInput)
	}
	if strings.TrimSpace(input.ClientID) == "" {
		return nil, nil, fmt.Errorf("%w: 缺少 client_id", ErrInvalidInput)
	}
	if strings.TrimSpace(input.RedirectURI) == "" {
		return nil, nil, fmt.Errorf("%w: 缺少 redirect_uri", ErrInvalidInput)
	}
	client, err := service.store.FindClientByClientID(ctx, strings.TrimSpace(input.ClientID))
	if err != nil {
		return nil, nil, err
	}
	if client == nil {
		return nil, nil, ErrInvalidClient
	}
	if strings.TrimSpace(input.CodeChallenge) == "" && strings.TrimSpace(input.CodeChallengeMethod) != "" {
		return nil, nil, fmt.Errorf("%w: 缺少 code_challenge", ErrInvalidInput)
	}
	if strings.TrimSpace(input.CodeChallenge) != "" {
		method := pkceMethod(input.CodeChallengeMethod)
		if method != "S256" && method != "plain" {
			return nil, nil, fmt.Errorf("%w: 仅支持 PKCE S256 或 plain", ErrInvalidInput)
		}
		input.CodeChallengeMethod = method
	}
	if !contains(client.RedirectURIs(), input.RedirectURI) {
		return nil, nil, fmt.Errorf("%w: redirect_uri 未在客户端白名单中", ErrInvalidInput)
	}
	allowedScopes, err := NormalizeKnownScopes(client.Scopes())
	if err != nil {
		return nil, nil, err
	}
	rawRequestedScopes := normalizeScopeList(strings.Fields(strings.ReplaceAll(strings.TrimSpace(input.Scope), ",", " ")))
	requestedScopes := allowedScopes
	if len(rawRequestedScopes) > 0 {
		requestedScopes, err = NormalizeKnownScopes(rawRequestedScopes)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, requestedScope := range requestedScopes {
		if len(allowedScopes) > 0 && !contains(allowedScopes, requestedScope) {
			return nil, nil, fmt.Errorf("%w: scope %s 未被客户端允许", ErrInvalidInput, requestedScope)
		}
	}
	return client, requestedScopes, nil
}

func (service *OAuthService) resolveAccessToken(ctx context.Context, rawToken string) (*model.AccessToken, *model.User, error) {
	token, err := service.store.FindAccessTokenByHash(ctx, security.HashToken(strings.TrimSpace(rawToken)))
	if err != nil {
		return nil, nil, err
	}
	if token == nil || token.RevokedAt != nil || token.ExpiresAt.Before(time.Now().UTC()) {
		return nil, nil, ErrInvalidToken
	}
	user, err := service.store.FindUserByID(ctx, token.UserID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidToken
	}
	return token, user, nil
}

func buildRedirect(rawURL string, values map[string]string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	for key, value := range values {
		query.Set(key, value)
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func DecodeBasicAuth(header string) (string, string, bool) {
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", false
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(header, "Basic "))
	if err != nil {
		return "", "", false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func normalizeScopeList(values []string) []string {
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

func hasScopeList(scopes []string, required string) bool {
	for _, scope := range scopes {
		if scope == required {
			return true
		}
	}
	return false
}

func pkceMethod(method string) string {
	trimmed := strings.TrimSpace(method)
	if trimmed == "" {
		return "plain"
	}
	return trimmed
}

func scopeKeys(definitions []ScopeDefinition) []string {
	items := make([]string, 0, len(definitions))
	for _, definition := range definitions {
		items = append(items, definition.Key)
	}
	return items
}
