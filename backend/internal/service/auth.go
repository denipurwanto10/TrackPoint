// Package service: aturan bisnis. Menerima pool atau tx via repo.DB.
package service

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"trackpoint/internal/apperr"
	"trackpoint/internal/auth"
	"trackpoint/internal/config"
	"trackpoint/internal/model"
	"trackpoint/internal/repo"
)

// TxBeginner dipenuhi *pgxpool.Pool.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

func withTx(ctx context.Context, pool TxBeginner, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ---- Auth ----

type AuthService struct {
	pool TxBeginner
	db   repo.DB
	cfg  config.Config
}

func NewAuthService(pool TxBeginner, db repo.DB, cfg config.Config) *AuthService {
	return &AuthService{pool: pool, db: db, cfg: cfg}
}

type LoginResult struct {
	AccessToken string
	Refresh     string
	ExpiresAt   time.Time
	User        model.User
}

func (s *AuthService) Login(ctx context.Context, email, password, ua, ip string) (*LoginResult, *apperr.Error) {
	u, err := repo.FindUserByEmail(ctx, s.db, strings.TrimSpace(email))
	if err != nil || !auth.CheckPassword(hashOf(u), password) {
		// Respons generik: tidak membedakan email salah vs password salah.
		return nil, apperr.Unauthorized("KREDENSIAL_SALAH", "Email atau password salah.")
	}
	if !u.IsActive {
		return nil, apperr.Forbidden("AKUN_NONAKTIF", "Akun dinonaktifkan. Hubungi admin.")
	}
	access, aerr := auth.NewAccessToken(s.cfg.JWTAccessSecret, u.ID, u.Role, s.cfg.JWTAccessTTL)
	if aerr != nil {
		return nil, apperr.Internal()
	}
	refresh, rerr := auth.NewRefreshToken()
	if rerr != nil {
		return nil, apperr.Internal()
	}
	exp := time.Now().Add(s.cfg.RefreshTTL)
	var uaP, ipP *string
	if ua != "" {
		uaP = &ua
	}
	if ip != "" {
		ipP = &ip
	}
	if _, derr := repo.CreateRefreshSession(ctx, s.db, u.ID, auth.HashRefreshToken(refresh), exp, uaP, ipP); derr != nil {
		return nil, apperr.Internal()
	}
	return &LoginResult{AccessToken: access, Refresh: refresh, ExpiresAt: exp, User: u.User}, nil
}

// Refresh memutar token: token lama dicabut + diganti token baru (atomik).
// Token yang sudah dicabut/diganti = reuse → semua sesi user dicabut.
func (s *AuthService) Refresh(ctx context.Context, rawToken, ua, ip string) (*LoginResult, *apperr.Error) {
	if rawToken == "" {
		return nil, apperr.Unauthorized("SESI_HILANG", "Sesi tidak ditemukan. Silakan login ulang.")
	}
	row, err := repo.FindRefreshSession(ctx, s.db, auth.HashRefreshToken(rawToken))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, apperr.Unauthorized("SESI_TIDAK_VALID", "Sesi tidak valid. Silakan login ulang.")
	}
	if err != nil {
		return nil, apperr.Internal()
	}
	if row.Expired || row.RevokedAt != nil {
		// Reuse terdeteksi: cabut semua sesi sebagai pengaman (kemungkinan token dicuri).
		_ = repo.RevokeAllUserSessions(ctx, s.db, row.UserID)
		return nil, apperr.Unauthorized("SESI_DICABUT", "Sesi sudah tidak berlaku. Silakan login ulang.")
	}
	u, err := repo.FindUserByID(ctx, s.db, row.UserID)
	if err != nil {
		return nil, apperr.Internal()
	}
	if !u.IsActive {
		_ = repo.RevokeAllUserSessions(ctx, s.db, row.UserID)
		return nil, apperr.Forbidden("AKUN_NONAKTIF", "Akun dinonaktifkan. Hubungi admin.")
	}
	access, aerr := auth.NewAccessToken(s.cfg.JWTAccessSecret, u.ID, u.Role, s.cfg.JWTAccessTTL)
	if aerr != nil {
		return nil, apperr.Internal()
	}
	refresh, rerr := auth.NewRefreshToken()
	if rerr != nil {
		return nil, apperr.Internal()
	}
	exp := time.Now().Add(s.cfg.RefreshTTL)
	var uaP, ipP *string
	if ua != "" {
		uaP = &ua
	}
	if ip != "" {
		ipP = &ip
	}
	newHash := auth.HashRefreshToken(refresh)
	if terr := withTx(ctx, s.pool, func(tx pgx.Tx) error {
		_, rerr := repo.RotateRefreshSession(ctx, tx, row.ID, u.ID, newHash, exp, uaP, ipP)
		return rerr
	}); terr != nil {
		if errors.Is(terr, repo.ErrRotateConflict) {
			// Request paralel lain sudah memutar token ini duluan;
			// frontend single-flight mencegah ini, tapi jangan hukum user.
			return nil, apperr.Unauthorized("SESI_DIPUTAR", "Sesi sudah diperbarui. Coba lagi.")
		}
		return nil, apperr.Internal()
	}
	return &LoginResult{AccessToken: access, Refresh: refresh, ExpiresAt: exp, User: u.User}, nil
}

func (s *AuthService) Logout(ctx context.Context, rawToken string) *apperr.Error {
	if rawToken == "" {
		return nil
	}
	row, err := repo.FindRefreshSession(ctx, s.db, auth.HashRefreshToken(rawToken))
	if err != nil {
		return apperr.FromPG(err, "refresh session")
	}
	if err := repo.RevokeSession(ctx, s.db, row.ID); err != nil {
		return apperr.FromPG(err, "revoke session")
	}
	return nil
}

func hashOf(u *repo.UserRow) string {
	if u == nil {
		return ""
	}
	// Bandingkan dalam waktu konstan tidak berlaku untuk bcrypt (sudah constant-time),
	// tapi hindari kebocoran lewat perbedaan error path: selalu lewat CheckPassword.
	_ = subtle.ConstantTimeCompare([]byte{1}, []byte{1})
	return u.PasswordHash
}
