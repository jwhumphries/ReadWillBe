package middleware

import (
	"net/http"

	"readwillbe/internal/model"

	"github.com/gorilla/sessions"
)

const (
	SessionKey             = "session"
	UserKey                = "session-user"
	SessionUserIDKey       = "userid"
	SessionLastSeenKey     = "last_seen"
	SessionRefreshInterval = 3600
)

func GetSecureSessionOptions(cfg model.Config) *sessions.Options {
	return &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24 hours
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteStrictMode,
	}
}
