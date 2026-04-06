package app

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/gofiber/fiber/v2"
)

//go:embed web-dist/* web-dist/assets/*
var embeddedFrontend embed.FS

func registerEmbeddedFrontend(app *fiber.App) error {
	frontendFS, err := fs.Sub(embeddedFrontend, "web-dist")
	if err != nil {
		return err
	}

	serve := func(c *fiber.Ctx) error {
		requestPath := strings.TrimPrefix(path.Clean(c.Path()), "/")
		if requestPath == "" || requestPath == "." {
			requestPath = "index.html"
		}

		if isReservedBackendPath(requestPath) {
			return fiber.ErrNotFound
		}

		if data, filePath, ok := readEmbeddedFile(frontendFS, requestPath); ok {
			setEmbeddedContentType(c, filePath, data)
			return c.Send(data)
		}

		if !strings.Contains(path.Base(requestPath), ".") {
			if data, filePath, ok := readEmbeddedFile(frontendFS, "index.html"); ok {
				setEmbeddedContentType(c, filePath, data)
				return c.Send(data)
			}
		}

		return fiber.ErrNotFound
	}

	app.Get("/", serve)
	app.Get("/*", serve)
	return nil
}

func isReservedBackendPath(requestPath string) bool {
	reservedPrefixes := []string{
		"api",
		"oauth",
		".well-known",
		"client-icons",
		"healthz",
	}
	for _, prefix := range reservedPrefixes {
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			return true
		}
	}
	return false
}

func readEmbeddedFile(frontendFS fs.FS, filePath string) ([]byte, string, bool) {
	data, err := fs.ReadFile(frontendFS, filePath)
	if err != nil {
		return nil, "", false
	}
	return data, filePath, true
}

func setEmbeddedContentType(c *fiber.Ctx, filePath string, data []byte) {
	contentType := mime.TypeByExtension(path.Ext(filePath))
	if contentType == "" && len(data) > 0 {
		contentType = http.DetectContentType(data)
	}
	if contentType != "" {
		c.Set(fiber.HeaderContentType, contentType)
	}
}
