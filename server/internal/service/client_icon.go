package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/model"
	"github.com/XianLinNet/XLNetAccount/internal/repository"
)

const maxClientIconSize = 4 << 20

var supportedClientIconTypes = map[string]string{
	"image/png":                ".png",
	"image/jpeg":               ".jpg",
	"image/gif":                ".gif",
	"image/webp":               ".webp",
	"image/svg+xml":            ".svg",
	"image/x-icon":             ".ico",
	"image/vnd.microsoft.icon": ".ico",
}

const (
	assetKeyClientIconPrefix = "client-icon-"
	assetKeyWebIcon          = "web-icon"
)

func NormalizeClientIconURL(ctx context.Context, store *repository.Store, clientID string, raw string) (string, error) {
	iconURL := strings.TrimSpace(raw)
	if iconURL == "" {
		if err := store.DeleteAsset(ctx, assetKeyClientIconPrefix+clientID); err != nil {
			return "", err
		}
		return "", nil
	}
	if isAssetIconURL(iconURL) {
		return iconURL, nil
	}
	parsed, err := url.Parse(iconURL)
	if err != nil {
		return "", fmt.Errorf("%w: 图标地址格式不正确", ErrInvalidInput)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return iconURL, nil
	}
	return downloadAndStoreClientIcon(ctx, store, clientID, iconURL)
}

func SyncClientIconURL(ctx context.Context, store *repository.Store, clientID string, raw string) (string, error) {
	iconURL := strings.TrimSpace(raw)
	if iconURL == "" || isAssetIconURL(iconURL) {
		return iconURL, nil
	}
	parsed, err := url.Parse(iconURL)
	if err != nil {
		return iconURL, nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return iconURL, nil
	}
	resolvedURL, err := downloadAndStoreClientIcon(ctx, store, clientID, iconURL)
	if err != nil {
		return iconURL, nil
	}
	return resolvedURL, nil
}

func NormalizeWebIconURL(ctx context.Context, store *repository.Store, raw string) (string, error) {
	iconURL := strings.TrimSpace(raw)
	if iconURL == "" {
		if err := store.DeleteAsset(ctx, assetKeyWebIcon); err != nil {
			return "", err
		}
		return "", nil
	}
	if isAssetIconURL(iconURL) {
		return iconURL, nil
	}
	parsed, err := url.Parse(iconURL)
	if err != nil {
		return "", fmt.Errorf("%w: 图标地址格式不正确", ErrInvalidInput)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return iconURL, nil
	}
	return downloadAndStoreWebIcon(ctx, store, iconURL)
}

func SyncWebIconURL(ctx context.Context, store *repository.Store, raw string) (string, error) {
	iconURL := strings.TrimSpace(raw)
	if iconURL == "" || isAssetIconURL(iconURL) {
		return iconURL, nil
	}
	parsed, err := url.Parse(iconURL)
	if err != nil {
		return iconURL, nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return iconURL, nil
	}
	resolvedURL, err := downloadAndStoreWebIcon(ctx, store, iconURL)
	if err != nil {
		return iconURL, nil
	}
	return resolvedURL, nil
}

func StoreUploadedClientIcon(ctx context.Context, store *repository.Store, fileHeader *multipart.FileHeader, clientID string) (string, error) {
	if fileHeader == nil {
		return "", fmt.Errorf("%w: 未找到上传文件", ErrInvalidInput)
	}
	data, mimeType, err := readAndValidateIcon(fileHeader)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	assetKey := assetKeyClientIconPrefix + clientID
	if err := store.SaveAsset(ctx, &model.Asset{
		Key:      assetKey,
		Data:     encoded,
		MimeType: mimeType,
	}); err != nil {
		return "", fmt.Errorf("保存图标失败: %w", err)
	}
	return "/api/assets/" + assetKey, nil
}

func StoreUploadedWebIcon(ctx context.Context, store *repository.Store, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", fmt.Errorf("%w: 未找到上传文件", ErrInvalidInput)
	}
	data, mimeType, err := readAndValidateIcon(fileHeader)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if err := store.SaveAsset(ctx, &model.Asset{
		Key:      assetKeyWebIcon,
		Data:     encoded,
		MimeType: mimeType,
	}); err != nil {
		return "", fmt.Errorf("保存图标失败: %w", err)
	}
	return "/api/assets/" + assetKeyWebIcon, nil
}

func DeleteClientIconAsset(ctx context.Context, store *repository.Store, clientID string) error {
	return store.DeleteAsset(ctx, assetKeyClientIconPrefix+clientID)
}

func downloadAndStoreClientIcon(ctx context.Context, store *repository.Store, clientID string, rawURL string) (string, error) {
	data, mimeType, err := downloadIcon(ctx, rawURL)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	assetKey := assetKeyClientIconPrefix + clientID
	if err := store.SaveAsset(ctx, &model.Asset{
		Key:      assetKey,
		Data:     encoded,
		MimeType: mimeType,
	}); err != nil {
		return "", fmt.Errorf("保存图标失败: %w", err)
	}
	return "/api/assets/" + assetKey, nil
}

func downloadAndStoreWebIcon(ctx context.Context, store *repository.Store, rawURL string) (string, error) {
	data, mimeType, err := downloadIcon(ctx, rawURL)
	if err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if err := store.SaveAsset(ctx, &model.Asset{
		Key:      assetKeyWebIcon,
		Data:     encoded,
		MimeType: mimeType,
	}); err != nil {
		return "", fmt.Errorf("保存图标失败: %w", err)
	}
	return "/api/assets/" + assetKeyWebIcon, nil
}

func downloadIcon(ctx context.Context, rawURL string) ([]byte, string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("%w: 图标请求地址无效", ErrInvalidInput)
	}
	request.Header.Set("Accept", "image/*")
	httpClient := &http.Client{Timeout: 15 * time.Second}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, "", fmt.Errorf("%w: 下载图标失败，请确认地址可访问", ErrInvalidInput)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("%w: 图标地址返回异常状态码 %d", ErrInvalidInput, response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxClientIconSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("%w: 读取图标内容失败", ErrInvalidInput)
	}
	mimeType := response.Header.Get("Content-Type")
	if _, err := validateIconData(data, mimeType); err != nil {
		return nil, "", err
	}
	return data, mimeType, nil
}

func readAndValidateIcon(fileHeader *multipart.FileHeader) ([]byte, string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, "", fmt.Errorf("%w: 读取上传文件失败", ErrInvalidInput)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxClientIconSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("%w: 读取图标内容失败", ErrInvalidInput)
	}
	contentType := fileHeader.Header.Get("Content-Type")
	extension, err := validateIconData(data, contentType)
	if err != nil {
		return nil, "", err
	}
	_ = extension
	return data, contentType, nil
}

func validateIconData(data []byte, contentTypeHeader string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("%w: 图标内容为空", ErrInvalidInput)
	}
	if len(data) > maxClientIconSize {
		return "", fmt.Errorf("%w: 图标文件过大，请控制在 4MB 以内", ErrInvalidInput)
	}
	contentType := strings.ToLower(strings.TrimSpace(contentTypeHeader))
	if contentType != "" {
		contentType, _, _ = mime.ParseMediaType(contentType)
	}
	if contentType == "" {
		contentType = http.DetectContentType(data)
	}
	extension, ok := supportedClientIconTypes[contentType]
	if !ok {
		return "", fmt.Errorf("%w: 不支持的图标类型 %s，仅支持 png/jpg/gif/webp/svg/ico", ErrInvalidInput, contentType)
	}
	return extension, nil
}

func isAssetIconURL(raw string) bool {
	return strings.HasPrefix(strings.TrimSpace(raw), "/api/assets/")
}
