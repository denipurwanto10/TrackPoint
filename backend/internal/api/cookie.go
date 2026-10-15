// Cookie refresh: HttpOnly, SameSite=Lax, Path /api/v1/auth, tanpa Domain.
// Secure mengikuti COOKIE_SECURE (false khusus HTTP localhost).
package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"trackpoint/internal/config"
)

const refreshCookieName = "tp_refresh"
const refreshCookiePath = "/api/v1/auth"

func setRefreshCookie(c *gin.Context, cfg config.Config, token string, maxAge int) {
	if maxAge <= 0 {
		maxAge = 30 * 24 * 3600
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    token,
		Path:     refreshCookiePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearRefreshCookie(c *gin.Context, cfg config.Config) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func nowUnix() int64 { return time.Now().Unix() }
