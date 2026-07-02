package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"strings"
	"time"

	"github.com/XianLinNet/XLNetAccount/internal/config"
	"github.com/XianLinNet/XLNetAccount/internal/service"
	"github.com/gofiber/fiber/v2"
)

type OAuthLoginHandler struct {
	oauthLoginService *service.OAuthLoginService
	cfg               config.Config
}

func NewOAuthLoginHandler(oauthLoginService *service.OAuthLoginService, cfg config.Config) *OAuthLoginHandler {
	return &OAuthLoginHandler{oauthLoginService: oauthLoginService, cfg: cfg}
}

// stateTTL is how long an OAuth state nonce is valid.
const stateTTL = 10 * time.Minute

// in-memory state store (single-instance only; for production use redis/db)
var oauthStates = struct {
	m map[string]time.Time
}{
	m: make(map[string]time.Time),
}

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			now := time.Now()
			for k, v := range oauthStates.m {
				if now.After(v) {
					delete(oauthStates.m, k)
				}
			}
		}
	}()
}

func generateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	state := hex.EncodeToString(b)
	oauthStates.m[state] = time.Now().Add(stateTTL)
	return state
}

func isValidState(state string) bool {
	expiry, ok := oauthStates.m[state]
	if !ok {
		return false
	}
	delete(oauthStates.m, state)
	return time.Now().Before(expiry)
}

// OAuthLogin redirects the user to the provider's OAuth consent page.
func (h *OAuthLoginHandler) OAuthLogin(c *fiber.Ctx) error {
	provider := service.OAuthProvider(strings.TrimSpace(c.Params("provider")))
	state := generateState()

	authURL, err := h.oauthLoginService.AuthorizeURL(provider, state)
	if err != nil {
		return writeError(c, fiber.StatusBadRequest, err.Error())
	}

	return c.Redirect(authURL, fiber.StatusFound)
}

// OAuthCallback handles the OAuth provider's callback, exchanges the code, and logs the user in.
func (h *OAuthLoginHandler) OAuthCallback(c *fiber.Ctx) error {
	provider := service.OAuthProvider(strings.TrimSpace(c.Params("provider")))
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	errorParam := strings.TrimSpace(c.Query("error"))

	if errorParam != "" {
		slog.Warn("oauth callback: provider returned error",
			slog.String("provider", string(provider)),
			slog.String("error", errorParam),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error="+errorParam, fiber.StatusFound)
	}

	if code == "" || !isValidState(state) {
		slog.Warn("oauth callback: invalid code or state",
			slog.String("provider", string(provider)),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=invalid_request", fiber.StatusFound)
	}

	info, err := h.oauthLoginService.ExchangeCode(context.Background(), provider, code)
	if err != nil {
		slog.Warn("oauth exchange failed",
			slog.String("provider", string(provider)),
			slog.String("error", err.Error()),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=exchange_failed", fiber.StatusFound)
	}

	user, rawToken, _, err := h.oauthLoginService.FindOrCreateUser(context.Background(), provider, info)
	if err != nil {
		slog.Error("oauth find or create user failed",
			slog.String("provider", string(provider)),
			slog.String("error", err.Error()),
		)
		return c.Redirect(h.cfg.WebBaseURL+"/auth/login?oauth_error=login_failed", fiber.StatusFound)
	}

	slog.Info("oauth login success",
		slog.String("provider", string(provider)),
		slog.String("username", user.Username),
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("ip", c.IP()),
	)

	// Redirect to frontend with session token in query (frontend picks it up)
	target := h.cfg.WebBaseURL + "/auth/login?oauth_token=" + rawToken + "&oauth_user=" + user.Username
	return c.Redirect(target, fiber.StatusFound)
}
