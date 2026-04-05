package service

import (
	"context"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xianlin-network/sso-platform/server/internal/config"
	"github.com/xianlin-network/sso-platform/server/internal/pkg/security"
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

func NormalizeClientIconURL(ctx context.Context, cfg config.Config, clientID string, raw string) (string, error) {
	iconURL := strings.TrimSpace(raw)
	if iconURL == "" {
		if err := removeClientIconFiles(cfg, clientID); err != nil {
			return "", err
		}
		return "", nil
	}
	if isUploadedClientIconURL(cfg, iconURL) {
		return adoptUploadedClientIcon(cfg, clientID, iconURL)
	}
	if isBackendClientIconURL(cfg, iconURL) {
		return iconURL, nil
	}
	parsed, err := url.Parse(iconURL)
	if err != nil {
		return "", fmt.Errorf("%w: 图标地址格式不正确", ErrInvalidInput)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return iconURL, nil
	}
	return downloadClientIcon(ctx, cfg, clientID, iconURL)
}

func SyncClientIconURL(ctx context.Context, cfg config.Config, clientID string, raw string) (string, error) {
	iconURL := strings.TrimSpace(raw)
	if iconURL == "" || isBackendClientIconURL(cfg, iconURL) {
		return iconURL, nil
	}
	parsed, err := url.Parse(iconURL)
	if err != nil {
		return iconURL, nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return iconURL, nil
	}
	resolvedURL, err := downloadClientIcon(ctx, cfg, clientID, iconURL)
	if err != nil {
		return iconURL, nil
	}
	return resolvedURL, nil
}

func ClientIconDir(cfg config.Config) string {
	return filepath.Join(cfg.AssetDir, "client-icons")
}

func StoreUploadedClientIcon(ctx context.Context, cfg config.Config, fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader == nil {
		return "", fmt.Errorf("%w: 未找到上传文件", ErrInvalidInput)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("%w: 读取上传文件失败", ErrInvalidInput)
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxClientIconSize+1))
	if err != nil {
		return "", fmt.Errorf("%w: 读取图标内容失败", ErrInvalidInput)
	}
	extension, err := validateClientIconData(data, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	uploadsDir := filepath.Join(ClientIconDir(cfg), "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return "", fmt.Errorf("准备图标目录失败: %w", err)
	}

	fileName := security.NewID() + extension
	filePath := filepath.Join(uploadsDir, fileName)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", fmt.Errorf("保存图标文件失败: %w", err)
	}

	return strings.TrimRight(cfg.ServerBaseURL, "/") + "/client-icons/uploads/" + fileName, nil
}

func downloadClientIcon(ctx context.Context, cfg config.Config, clientID string, rawURL string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", fmt.Errorf("%w: 图标请求地址无效", ErrInvalidInput)
	}
	request.Header.Set("Accept", "image/*")

	httpClient := &http.Client{Timeout: 15 * time.Second}
	response, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("%w: 下载图标失败，请确认地址可访问", ErrInvalidInput)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("%w: 图标地址返回异常状态码 %d", ErrInvalidInput, response.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, maxClientIconSize+1))
	if err != nil {
		return "", fmt.Errorf("%w: 读取图标内容失败", ErrInvalidInput)
	}
	extension, err := validateClientIconData(data, response.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(ClientIconDir(cfg), 0o755); err != nil {
		return "", fmt.Errorf("准备图标目录失败: %w", err)
	}
	if err := removeClientIconFiles(cfg, clientID); err != nil {
		return "", err
	}

	fileName := clientID + extension
	filePath := filepath.Join(ClientIconDir(cfg), fileName)
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return "", fmt.Errorf("保存图标文件失败: %w", err)
	}

	return strings.TrimRight(cfg.ServerBaseURL, "/") + "/client-icons/" + fileName, nil
}

func validateClientIconData(data []byte, contentTypeHeader string) (string, error) {
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

func adoptUploadedClientIcon(cfg config.Config, clientID string, raw string) (string, error) {
	uploadURLPrefix := strings.TrimRight(cfg.ServerBaseURL, "/") + "/client-icons/uploads/"
	fileName := strings.TrimPrefix(strings.TrimSpace(raw), uploadURLPrefix)
	if fileName == strings.TrimSpace(raw) || strings.Contains(fileName, "/") || strings.Contains(fileName, `\`) {
		return "", fmt.Errorf("%w: 上传图标地址无效", ErrInvalidInput)
	}
	if err := removeClientIconFiles(cfg, clientID); err != nil {
		return "", err
	}
	uploadsDir := filepath.Join(ClientIconDir(cfg), "uploads")
	sourcePath := filepath.Join(uploadsDir, fileName)
	extension := filepath.Ext(fileName)
	if extension == "" {
		return "", fmt.Errorf("%w: 上传图标文件无扩展名", ErrInvalidInput)
	}
	targetName := clientID + extension
	targetPath := filepath.Join(ClientIconDir(cfg), targetName)
	if err := os.MkdirAll(ClientIconDir(cfg), 0o755); err != nil {
		return "", fmt.Errorf("准备图标目录失败: %w", err)
	}
	if err := os.Rename(sourcePath, targetPath); err != nil {
		return "", fmt.Errorf("保存图标文件失败: %w", err)
	}
	return strings.TrimRight(cfg.ServerBaseURL, "/") + "/client-icons/" + targetName, nil
}

func removeClientIconFiles(cfg config.Config, clientID string) error {
	entries, err := os.ReadDir(ClientIconDir(cfg))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取图标目录失败: %w", err)
	}
	prefix := clientID + "."
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) {
			continue
		}
		if err := os.Remove(filepath.Join(ClientIconDir(cfg), entry.Name())); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("删除旧图标文件失败: %w", err)
		}
	}
	return nil
}

func DeleteClientIconFiles(cfg config.Config, clientID string) error {
	return removeClientIconFiles(cfg, clientID)
}

func isBackendClientIconURL(cfg config.Config, raw string) bool {
	prefix := strings.TrimRight(cfg.ServerBaseURL, "/") + "/client-icons/"
	trimmed := strings.TrimSpace(raw)
	return strings.HasPrefix(trimmed, prefix) && !isUploadedClientIconURL(cfg, trimmed)
}

func isUploadedClientIconURL(cfg config.Config, raw string) bool {
	prefix := strings.TrimRight(cfg.ServerBaseURL, "/") + "/client-icons/uploads/"
	return strings.HasPrefix(strings.TrimSpace(raw), prefix)
}
