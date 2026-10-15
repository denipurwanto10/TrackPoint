// Package mw: autentikasi Bearer JWT + RBAC + origin check untuk cookie.
package mw

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"trackpoint/internal/auth"
	"trackpoint/internal/config"
	"trackpoint/internal/model"
	"trackpoint/internal/repo"
)

// Auth memvalidasi Bearer token, memuat user aktif, menyimpan di context.
// Gagal → 401 generik tanpa detail internal.
func Auth(cfg config.Config, db repo.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		tok := ""
		if strings.HasPrefix(h, "Bearer ") {
			tok = strings.TrimPrefix(h, "Bearer ")
		}
		if tok == "" {
			abort(c, http.StatusUnauthorized, "TIDAK_TERAUTENTIKASI", "Silakan login dulu.")
			return
		}
		claims, err := auth.ParseAccessToken(cfg.JWTAccessSecret, tok)
		if err != nil {
			if err.Error() == "expired" {
				abort(c, http.StatusUnauthorized, "SESI_KEDALUWARSA", "Sesi kedaluwarsa. Silakan refresh atau login ulang.")
			} else {
				abort(c, http.StatusUnauthorized, "TOKEN_TIDAK_VALID", "Token tidak valid. Silakan login ulang.")
			}
			return
		}
		u, derr := repo.FindUserByID(c.Request.Context(), db, claims.UserID)
		if derr != nil {
			abort(c, http.StatusUnauthorized, "TIDAK_TERAUTENTIKASI", "Silakan login dulu.")
			return
		}
		if !u.IsActive {
			abort(c, http.StatusForbidden, "AKUN_NONAKTIF", "Akun dinonaktifkan. Hubungi admin.")
			return
		}
		c.Set("user", model.ContextUser{
			ID: u.ID, Role: u.Role, Nama: u.Nama, Email: u.Email, IsActive: u.IsActive,
		})
		c.Next()
	}
}

// RequireRole: RBAC backend. Frontend boleh menyembunyikan menu, tapi izin final di sini.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := map[string]bool{}
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		u, ok := Current(c)
		if !ok || !allowed[u.Role] {
			abort(c, http.StatusForbidden, "AKSES_DITOLAK", "Anda tidak punya akses ke fitur ini.")
			return
		}
		c.Next()
	}
}

// Current mengambil user dari context.
func Current(c *gin.Context) (model.ContextUser, bool) {
	v, ok := c.Get("user")
	if !ok {
		return model.ContextUser{}, false
	}
	u, ok := v.(model.ContextUser)
	return u, ok
}

// CheckOrigin untuk endpoint cookie (refresh/logout): tolak cross-site tanpa Origin/Referer yang cocok.
// Klien non-browser tanpa Origin/Referer tetap diizinkan (curl, SSR) — cookie sendiri yang jadi bukti.
func CheckOrigin(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		ref := c.GetHeader("Referer")
		if origin == "" && ref == "" {
			c.Next()
			return
		}
		if origin != "" && !sameSite(cfg.CORSOrigin, origin) {
			abort(c, http.StatusForbidden, "ORIGIN_DITOLAK", "Origin tidak diizinkan.")
			return
		}
		if origin == "" && ref != "" && !refStartsWith(ref, cfg.CORSOrigin) {
			abort(c, http.StatusForbidden, "ORIGIN_DITOLAK", "Origin tidak diizinkan.")
			return
		}
		c.Next()
	}
}

func sameSite(a, b string) bool {
	return strings.TrimSuffix(a, "/") == strings.TrimSuffix(b, "/")
}

func refStartsWith(ref, origin string) bool {
	return strings.HasPrefix(ref, strings.TrimSuffix(origin, "/"))
}

func abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}
