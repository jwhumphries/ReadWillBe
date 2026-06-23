package middleware

import (
	"net/http"

	"readwillbe/internal/model"

	"github.com/gorilla/sessions"
)

// Session and context keys used by the middleware package.
const (
	SessionKey             = "session"
	UserKey                = "session-user"
	SessionUserIDKey       = "userid"
	SessionLastSeenKey     = "last_seen"
	SessionRefreshInterval = 3600
)

// GetSecureSessionOptions returns gorilla/sessions options with secure defaults
// (HttpOnly, SameSite=Strict, Secure in production) for the given config.
func GetSecureSessionOptions(cfg model.Config) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteStrictMode,
	}
}
