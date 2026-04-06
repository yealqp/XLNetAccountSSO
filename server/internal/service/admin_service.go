package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

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
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type UpdateUserInput struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Password string `json:"password"`
}

type AdminService struct {
	store *repository.Store
	cfg   config.Config
}

const (
	settingPlatformName      = "platform_name"
	settingAllowRegistration = "allow_registration"
	settingSMTPHost          = "smtp_host"
	settingSMTPUser          = "smtp_user"
	settingSMTPPassword      = "smtp_password"
	settingSMTPPort          = "smtp_port"
	settingSMTPTLS           = "smtp_tls"
	settingCAPAPIEndpoint    = "cap_api_endpoint"
	settingCAPSecretKey      = "cap_secret_key"
)

type PlatformSettings struct {
	PlatformName      string `json:"platform_name"`
	AllowRegistration bool   `json:"allow_registration"`
	SMTPHost          string `json:"smtp_host"`
	SMTPUser          string `json:"smtp_user"`
	SMTPPassword      string `json:"smtp_password"`
	SMTPPort          string `json:"smtp_port"`
	SMTPTLS           bool   `json:"smtp_tls"`
	CAPAPIEndpoint    string `json:"cap_api_endpoint"`
	CAPSecretKey      string `json:"cap_secret_key"`
}

type TestEmailInput struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	SMTPPort     string `json:"smtp_port"`
	SMTPTLS      bool   `json:"smtp_tls"`
	To           string `json:"to"`
}

func NewAdminService(store *repository.Store, cfg config.Config) *AdminService {
	return &AdminService{store: store, cfg: cfg}
}

func (service *AdminService) EnsureDefaults(ctx context.Context) error {
	defaults := map[string]string{
		settingPlatformName:      strings.TrimSpace(service.cfg.AppName),
		settingAllowRegistration: "false",
		settingSMTPPort:          "587",
		settingSMTPTLS:           "true",
	}
	for key, value := range defaults {
		setting, err := service.store.FindPlatformSetting(ctx, key)
		if err != nil {
			return err
		}
		if setting != nil && strings.TrimSpace(setting.Value) != "" {
			continue
		}
		if err := service.store.SavePlatformSetting(ctx, &model.PlatformSetting{Key: key, Value: value}); err != nil {
			return err
		}
	}
	return nil
}

func (service *AdminService) PublicSettings(ctx context.Context) (map[string]any, error) {
	settings, err := service.loadPlatformSettings(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"platform_name":      settings.PlatformName,
		"allow_registration": settings.AllowRegistration,
		"cap_api_endpoint":   settings.CAPAPIEndpoint,
	}, nil
}

func (service *AdminService) PlatformName(ctx context.Context) (string, error) {
	settings, err := service.loadPlatformSettings(ctx)
	if err != nil {
		return "", err
	}
	return settings.PlatformName, nil
}

func (service *AdminService) PlatformSettings(ctx context.Context) (PlatformSettings, error) {
	return service.loadPlatformSettings(ctx)
}

func (service *AdminService) UpdatePlatformSettings(ctx context.Context, input PlatformSettings) (PlatformSettings, error) {
	platformName := strings.TrimSpace(input.PlatformName)
	if platformName == "" {
		return PlatformSettings{}, fmt.Errorf("%w: 平台名称不能为空", ErrInvalidInput)
	}
	settings := map[string]string{
		settingPlatformName:      platformName,
		settingAllowRegistration: boolString(input.AllowRegistration),
		settingSMTPHost:          strings.TrimSpace(input.SMTPHost),
		settingSMTPUser:          strings.TrimSpace(input.SMTPUser),
		settingSMTPPassword:      strings.TrimSpace(input.SMTPPassword),
		settingSMTPPort:          strings.TrimSpace(input.SMTPPort),
		settingSMTPTLS:           boolString(input.SMTPTLS),
		settingCAPAPIEndpoint:    strings.TrimSpace(input.CAPAPIEndpoint),
		settingCAPSecretKey:      strings.TrimSpace(input.CAPSecretKey),
	}
	for key, value := range settings {
		if err := service.store.SavePlatformSetting(ctx, &model.PlatformSetting{Key: key, Value: value}); err != nil {
			return PlatformSettings{}, err
		}
	}
	return service.loadPlatformSettings(ctx)
}

func (service *AdminService) SendTestEmail(ctx context.Context, input TestEmailInput) error {
	settings := PlatformSettings{
		SMTPHost:     strings.TrimSpace(input.SMTPHost),
		SMTPUser:     strings.TrimSpace(input.SMTPUser),
		SMTPPassword: strings.TrimSpace(input.SMTPPassword),
		SMTPPort:     strings.TrimSpace(input.SMTPPort),
		SMTPTLS:      input.SMTPTLS,
	}
	to := strings.TrimSpace(strings.ToLower(input.To))
	if to == "" || !strings.Contains(to, "@") {
		return fmt.Errorf("%w: 请输入有效的测试收件邮箱", ErrInvalidInput)
	}
	platformName, err := service.PlatformName(ctx)
	if err != nil {
		return err
	}
	subject := platformName + " 邮件测试"
	body := "这是一封来自 " + platformName + " 的测试邮件。\n\n如果您收到此邮件，说明当前 SMTP 配置可正常发送。"
	return sendSMTPMail(settings, to, subject, body)
}

func (service *AdminService) Overview(ctx context.Context) (repository.Overview, error) {
	return service.store.Overview(ctx)
}

func (service *AdminService) OverviewForUser(ctx context.Context, actor *model.User) (repository.Overview, error) {
	var result repository.Overview
	if actor == nil {
		return result, ErrUnauthorized
	}
	clients, err := service.store.CountClientsByCreator(ctx, actor.ID)
	if err != nil {
		return result, err
	}
	accessTokens, err := service.store.ListAccessTokensByUser(ctx, actor.ID)
	if err != nil {
		return result, err
	}
	activeTokens := int64(0)
	now := time.Now().UTC()
	for _, token := range accessTokens {
		if token.RevokedAt == nil && token.ExpiresAt.After(now) {
			activeTokens++
		}
	}
	result.Users = 1
	result.Clients = clients
	result.AccessTokens = activeTokens
	result.ActiveSessions = 0
	return result, nil
}

func (service *AdminService) ListClients(ctx context.Context) ([]map[string]any, error) {
	clients, err := service.store.ListClients(ctx)
	if err != nil {
		return nil, err
	}
	ownerNames, err := service.usernamesByID(ctx, collectClientOwnerIDs(clients))
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
		items = append(items, clientResponse(client, ownerNames[client.CreatedBy], ""))
	}
	return items, nil
}

func (service *AdminService) ListUserClients(ctx context.Context, actor *model.User) ([]map[string]any, error) {
	if actor == nil {
		return nil, ErrUnauthorized
	}
	clients, err := service.store.ListClientsByCreator(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	ownerNames, err := service.usernamesByID(ctx, collectClientOwnerIDs(clients))
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
		items = append(items, clientResponse(client, ownerNames[client.CreatedBy], ""))
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

	return clientResponse(*client, actor.Username, rawSecret), nil
}

func (service *AdminService) UpdateClient(ctx context.Context, id string, input UpdateClientInput) (map[string]any, error) {
	return service.updateClient(ctx, nil, id, input, true)
}

func (service *AdminService) UpdateUserClient(ctx context.Context, actor *model.User, id string, input UpdateClientInput) (map[string]any, error) {
	return service.updateClient(ctx, actor, id, input, false)
}

func (service *AdminService) updateClient(ctx context.Context, actor *model.User, id string, input UpdateClientInput, manageAll bool) (map[string]any, error) {
	client, err := service.store.FindClientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, ErrNotFound
	}
	if !manageAll {
		if actor == nil || client.CreatedBy != actor.ID {
			return nil, ErrForbidden
		}
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
	ownerName := ""
	if owner, err := service.store.FindUserByID(ctx, client.CreatedBy); err == nil && owner != nil {
		ownerName = owner.Username
	}
	return clientResponse(*client, ownerName, ""), nil
}

func (service *AdminService) DeleteClient(ctx context.Context, id string) error {
	return service.deleteClient(ctx, nil, id, true)
}

func (service *AdminService) DeleteUserClient(ctx context.Context, actor *model.User, id string) error {
	return service.deleteClient(ctx, actor, id, false)
}

func (service *AdminService) deleteClient(ctx context.Context, actor *model.User, id string, manageAll bool) error {
	client, err := service.store.FindClientByID(ctx, id)
	if err != nil {
		return err
	}
	if client == nil {
		return ErrNotFound
	}
	if !manageAll {
		if actor == nil || client.CreatedBy != actor.ID {
			return ErrForbidden
		}
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
	email := strings.TrimSpace(strings.ToLower(input.Email))
	existing, err := service.store.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrConflict
	}
	if email != "" {
		existingByEmail, err := service.store.FindUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingByEmail != nil {
			return nil, ErrConflict
		}
	}
	if err := validateRegistrationPassword(input.Password); err != nil {
		return nil, err
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
		Username:     username,
		PasswordHash: passwordHash,
		Email:        email,
		Role:         role,
		Status:       "active",
	}
	if err := service.store.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return userResponse(*user), nil
}

func (service *AdminService) UpdateUser(ctx context.Context, id string, input UpdateUserInput) (map[string]any, error) {
	userID, err := parseUintID(id)
	if err != nil {
		return nil, err
	}
	user, err := service.store.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}
	if username := strings.TrimSpace(input.Username); username != "" {
		existingByUsername, err := service.store.FindUserByUsername(ctx, username)
		if err != nil {
			return nil, err
		}
		if existingByUsername != nil && existingByUsername.ID != user.ID {
			return nil, ErrConflict
		}
		user.Username = username
	}
	if email := strings.TrimSpace(strings.ToLower(input.Email)); email != "" {
		existingByEmail, err := service.store.FindUserByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingByEmail != nil && existingByEmail.ID != user.ID {
			return nil, ErrConflict
		}
		user.Email = email
	}
	if role := strings.TrimSpace(input.Role); role != "" {
		user.Role = role
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		user.Status = status
	}
	if password := strings.TrimSpace(input.Password); password != "" {
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
	return userResponse(*user), nil
}

func (service *AdminService) DeleteUser(ctx context.Context, actor *model.User, id string) error {
	userID, err := parseUintID(id)
	if err != nil {
		return err
	}
	user, err := service.store.FindUserByID(ctx, userID)
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
	return service.store.DeleteUser(ctx, userID)
}

func (service *AdminService) loadPlatformSettings(ctx context.Context) (PlatformSettings, error) {
	platformName, err := service.settingValue(ctx, settingPlatformName, strings.TrimSpace(service.cfg.AppName))
	if err != nil {
		return PlatformSettings{}, err
	}
	allowRegistration, err := service.settingValue(ctx, settingAllowRegistration, "false")
	if err != nil {
		return PlatformSettings{}, err
	}
	smtpHost, err := service.settingValue(ctx, settingSMTPHost, "")
	if err != nil {
		return PlatformSettings{}, err
	}
	smtpUser, err := service.settingValue(ctx, settingSMTPUser, "")
	if err != nil {
		return PlatformSettings{}, err
	}
	smtpPassword, err := service.settingValue(ctx, settingSMTPPassword, "")
	if err != nil {
		return PlatformSettings{}, err
	}
	smtpPort, err := service.settingValue(ctx, settingSMTPPort, "587")
	if err != nil {
		return PlatformSettings{}, err
	}
	smtpTLS, err := service.settingValue(ctx, settingSMTPTLS, "true")
	if err != nil {
		return PlatformSettings{}, err
	}
	capAPIEndpoint, err := service.settingValue(ctx, settingCAPAPIEndpoint, "")
	if err != nil {
		return PlatformSettings{}, err
	}
	capSecretKey, err := service.settingValue(ctx, settingCAPSecretKey, "")
	if err != nil {
		return PlatformSettings{}, err
	}
	return PlatformSettings{
		PlatformName:      platformName,
		AllowRegistration: parseBoolString(allowRegistration),
		SMTPHost:          smtpHost,
		SMTPUser:          smtpUser,
		SMTPPassword:      smtpPassword,
		SMTPPort:          smtpPort,
		SMTPTLS:           parseBoolString(smtpTLS),
		CAPAPIEndpoint:    capAPIEndpoint,
		CAPSecretKey:      capSecretKey,
	}, nil
}

func (service *AdminService) settingValue(ctx context.Context, key string, fallback string) (string, error) {
	setting, err := service.store.FindPlatformSetting(ctx, key)
	if err != nil {
		return "", err
	}
	if setting == nil || strings.TrimSpace(setting.Value) == "" {
		return fallback, nil
	}
	return strings.TrimSpace(setting.Value), nil
}

func parseBoolString(value string) bool {
	return strings.EqualFold(strings.TrimSpace(value), "true") || strings.TrimSpace(value) == "1"
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func clientResponse(client model.OAuthClient, ownerUsername string, rawSecret string) map[string]any {
	return map[string]any{
		"id":             client.ID,
		"name":           client.Name,
		"description":    client.Description,
		"icon_url":       client.IconURL,
		"client_id":      client.ClientID,
		"client_type":    client.ClientType,
		"redirect_uris":  client.RedirectURIs(),
		"scopes":         client.Scopes(),
		"trusted":        client.Trusted,
		"created_by":     client.CreatedBy,
		"owner_username": ownerUsername,
		"created_at":     client.CreatedAt,
		"updated_at":     client.UpdatedAt,
		"client_secret":  rawSecret,
	}
}

func userResponse(user model.User) map[string]any {
	return map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"status":     user.Status,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

func parseUintID(value string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: 无效的用户 ID", ErrInvalidInput)
	}
	return uint(parsed), nil
}

func (service *AdminService) usernamesByID(ctx context.Context, ids []uint) (map[uint]string, error) {
	result := map[uint]string{}
	for _, id := range ids {
		if _, ok := result[id]; ok {
			continue
		}
		user, err := service.store.FindUserByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if user != nil {
			result[id] = user.Username
		}
	}
	return result, nil
}

func collectClientOwnerIDs(clients []model.OAuthClient) []uint {
	ids := make([]uint, 0, len(clients))
	seen := map[uint]struct{}{}
	for _, client := range clients {
		if _, ok := seen[client.CreatedBy]; ok {
			continue
		}
		seen[client.CreatedBy] = struct{}{}
		ids = append(ids, client.CreatedBy)
	}
	return ids
}
