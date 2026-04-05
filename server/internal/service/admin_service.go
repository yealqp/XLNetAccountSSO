package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/xianlin-network/sso-platform/server/internal/config"
	"github.com/xianlin-network/sso-platform/server/internal/model"
	"github.com/xianlin-network/sso-platform/server/internal/pkg/security"
	"github.com/xianlin-network/sso-platform/server/internal/repository"
)

type CreateClientInput struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	IconURL      string   `json:"icon_url"`
	ClientID     string   `json:"client_id"`
	ClientType   string   `json:"client_type"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	Trusted      bool     `json:"trusted"`
}

type UpdateClientInput struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	IconURL      string   `json:"icon_url"`
	RedirectURIs []string `json:"redirect_uris"`
	Scopes       []string `json:"scopes"`
	Trusted      bool     `json:"trusted"`
}

type CreateUserInput struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
}

type UpdateUserInput struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	Password    string `json:"password"`
}

type AdminService struct {
	store *repository.Store
	cfg   config.Config
}

const settingPlatformName = "platform_name"

func NewAdminService(store *repository.Store, cfg config.Config) *AdminService {
	return &AdminService{store: store, cfg: cfg}
}

func (service *AdminService) EnsureDefaults(ctx context.Context) error {
	setting, err := service.store.FindPlatformSetting(ctx, settingPlatformName)
	if err != nil {
		return err
	}
	if setting != nil && strings.TrimSpace(setting.Value) != "" {
		return nil
	}
	return service.store.SavePlatformSetting(ctx, &model.PlatformSetting{
		Key:   settingPlatformName,
		Value: strings.TrimSpace(service.cfg.AppName),
	})
}

func (service *AdminService) PublicSettings(ctx context.Context) (map[string]any, error) {
	platformName, err := service.PlatformName(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"platform_name": platformName,
	}, nil
}

func (service *AdminService) PlatformName(ctx context.Context) (string, error) {
	setting, err := service.store.FindPlatformSetting(ctx, settingPlatformName)
	if err != nil {
		return "", err
	}
	if setting == nil || strings.TrimSpace(setting.Value) == "" {
		return strings.TrimSpace(service.cfg.AppName), nil
	}
	return strings.TrimSpace(setting.Value), nil
}

func (service *AdminService) UpdatePlatformName(ctx context.Context, value string) (map[string]any, error) {
	platformName := strings.TrimSpace(value)
	if platformName == "" {
		return nil, fmt.Errorf("%w: 平台名称不能为空", ErrInvalidInput)
	}
	setting := &model.PlatformSetting{
		Key:   settingPlatformName,
		Value: platformName,
	}
	if err := service.store.SavePlatformSetting(ctx, setting); err != nil {
		return nil, err
	}
	return map[string]any{
		"platform_name": platformName,
	}, nil
}

func (service *AdminService) Overview(ctx context.Context) (repository.Overview, error) {
	return service.store.Overview(ctx)
}

func (service *AdminService) ListClients(ctx context.Context) ([]map[string]any, error) {
	clients, err := service.store.ListClients(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(clients))
	for index := range clients {
		client := clients[index]
		resolvedIconURL, err := SyncClientIconURL(ctx, service.cfg, client.ID, client.IconURL)
		if err == nil && resolvedIconURL != client.IconURL {
			client.IconURL = resolvedIconURL
			_ = service.store.SaveClient(ctx, &client)
		}
		items = append(items, clientResponse(client, ""))
	}
	return items, nil
}

func (service *AdminService) UploadClientIcon(ctx context.Context, fileHeader *multipart.FileHeader) (map[string]any, error) {
	iconURL, err := StoreUploadedClientIcon(ctx, service.cfg, fileHeader)
	if err != nil {
		return nil, err
	}
	return map[string]any{"icon_url": iconURL}, nil
}

func (service *AdminService) CreateClient(ctx context.Context, actor *model.User, input CreateClientInput) (map[string]any, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" || len(input.RedirectURIs) == 0 {
		return nil, ErrInvalidInput
	}
	scopes, err := NormalizeKnownScopes(input.Scopes)
	if err != nil {
		return nil, err
	}
	clientType := strings.TrimSpace(input.ClientType)
	if clientType == "" {
		clientType = "public"
	}
	if clientType != "public" && clientType != "confidential" {
		return nil, ErrInvalidInput
	}
	clientID := strings.TrimSpace(input.ClientID)
	if clientID == "" {
		clientID = fmt.Sprintf("client_%s", security.NewID()[:12])
	}
	existing, err := service.store.FindClientByClientID(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrConflict
	}

	if actor == nil {
		return nil, ErrUnauthorized
	}

	client := &model.OAuthClient{
		ID:          security.NewID(),
		Name:        name,
		Description: strings.TrimSpace(input.Description),
		ClientID:    clientID,
		ClientType:  clientType,
		Trusted:     input.Trusted,
		CreatedBy:   actor.ID,
	}
	iconURL, err := NormalizeClientIconURL(ctx, service.cfg, client.ID, input.IconURL)
	if err != nil {
		return nil, err
	}
	client.IconURL = iconURL
	client.SetRedirectURIs(input.RedirectURIs)
	client.SetScopes(scopes)

	rawSecret := ""
	if clientType == "confidential" {
		rawSecret, err = security.NewOpaqueToken(24)
		if err != nil {
			return nil, err
		}
		hash, err := security.HashPassword(rawSecret)
		if err != nil {
			return nil, err
		}
		client.ClientSecretHash = hash
	}

	if err := service.store.CreateClient(ctx, client); err != nil {
		return nil, err
	}

	return clientResponse(*client, rawSecret), nil
}

func (service *AdminService) UpdateClient(ctx context.Context, id string, input UpdateClientInput) (map[string]any, error) {
	client, err := service.store.FindClientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrNotFound
	}
	name := strings.TrimSpace(input.Name)
	if name == "" || len(input.RedirectURIs) == 0 {
		return nil, ErrInvalidInput
	}
	scopes, err := NormalizeKnownScopes(input.Scopes)
	if err != nil {
		return nil, err
	}
	client.Name = name
	client.Description = strings.TrimSpace(input.Description)
	iconURL, err := NormalizeClientIconURL(ctx, service.cfg, client.ID, input.IconURL)
	if err != nil {
		return nil, err
	}
	client.IconURL = iconURL
	client.Trusted = input.Trusted
	client.SetRedirectURIs(input.RedirectURIs)
	client.SetScopes(scopes)
	if err := service.store.SaveClient(ctx, client); err != nil {
		return nil, err
	}
	return clientResponse(*client, ""), nil
}

func (service *AdminService) DeleteClient(ctx context.Context, id string) error {
	client, err := service.store.FindClientByID(ctx, id)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrNotFound
	}
	if err := service.store.DeleteAuthorizationCodesByClientID(ctx, client.ClientID); err != nil {
		return err
	}
	if err := service.store.DeleteRefreshTokensByClientID(ctx, client.ClientID); err != nil {
		return err
	}
	if err := service.store.DeleteAccessTokensByClientID(ctx, client.ClientID); err != nil {
		return err
	}
	if err := service.store.DeleteClient(ctx, id); err != nil {
		return err
	}
	if err := DeleteClientIconFiles(service.cfg, client.ID); err != nil {
		return err
	}
	return nil
}

func (service *AdminService) ListUsers(ctx context.Context) ([]map[string]any, error) {
	users, err := service.store.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(users))
	for _, user := range users {
		items = append(items, userResponse(user))
	}
	return items, nil
}

func (service *AdminService) CreateUser(ctx context.Context, input CreateUserInput) (map[string]any, error) {
	username := strings.TrimSpace(input.Username)
	if username == "" || strings.TrimSpace(input.Password) == "" {
		return nil, ErrInvalidInput
	}
	existing, err := service.store.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrConflict
	}
	passwordHash, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	role := strings.TrimSpace(input.Role)
	if role == "" {
		role = "user"
	}
	user := &model.User{
		ID:           security.NewID(),
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  strings.TrimSpace(input.DisplayName),
		Email:        strings.TrimSpace(input.Email),
		Role:         role,
		Status:       "active",
	}
	if user.DisplayName == "" {
		user.DisplayName = username
	}
	if err := service.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return userResponse(*user), nil
}

func (service *AdminService) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (map[string]any, error) {
	user, err := service.store.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	if displayName := strings.TrimSpace(input.DisplayName); displayName != "" {
		user.DisplayName = displayName
	}
	if email := strings.TrimSpace(input.Email); email != "" {
		user.Email = email
	}
	if role := strings.TrimSpace(input.Role); role != "" {
		user.Role = role
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		user.Status = status
	}
	if password := strings.TrimSpace(input.Password); password != "" {
		hash, err := security.HashPassword(password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	if err := service.store.SaveUser(ctx, user); err != nil {
		return nil, err
	}
	return userResponse(*user), nil
}

func (service *AdminService) DeleteUser(ctx context.Context, actor *model.User, id string) error {
	user, err := service.store.FindUserByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotFound
	}
	if actor != nil && actor.ID == user.ID {
		return fmt.Errorf("%w: 不能删除当前登录用户", ErrInvalidInput)
	}
	if err := service.store.DeleteAuthorizationCodesByUserID(ctx, user.ID); err != nil {
		return err
	}
	if err := service.store.DeleteRefreshTokensByUserID(ctx, user.ID); err != nil {
		return err
	}
	if err := service.store.DeleteAccessTokensByUserID(ctx, user.ID); err != nil {
		return err
	}
	if err := service.store.DeleteSessionsByUserID(ctx, user.ID); err != nil {
		return err
	}
	return service.store.DeleteUser(ctx, id)
}

func clientResponse(client model.OAuthClient, rawSecret string) map[string]any {
	return map[string]any{
		"id":            client.ID,
		"name":          client.Name,
		"description":   client.Description,
		"icon_url":      client.IconURL,
		"client_id":     client.ClientID,
		"client_type":   client.ClientType,
		"redirect_uris": client.RedirectURIs(),
		"scopes":        client.Scopes(),
		"trusted":       client.Trusted,
		"created_by":    client.CreatedBy,
		"created_at":    client.CreatedAt,
		"updated_at":    client.UpdatedAt,
		"client_secret": rawSecret,
	}
}

func userResponse(user model.User) map[string]any {
	return map[string]any{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
		"email":        user.Email,
		"role":         user.Role,
		"status":       user.Status,
		"created_at":   user.CreatedAt,
		"updated_at":   user.UpdatedAt,
	}
}
