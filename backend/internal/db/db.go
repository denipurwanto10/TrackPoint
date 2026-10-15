// Package db menyediakan koneksi PostgreSQL via pgxpool dengan inisialisasi,
// penutupan graceful, dan error yang tidak membocorkan secret.
package db

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config holder untuk konfigurasi pool DB.
type Config struct {
	URL         string
	MaxOpenCon  int
	MaxIdleConn int // tanpa padanan di pgxpool v5: sengaja tidak diterapkan (lihat NewPool)
	IdleTimeout time.Duration
}

// DefaultConfig mengembalikan konfigurasi pool standar.
func DefaultConfig() Config {
	return Config{
		MaxOpenCon:  25,
		MaxIdleConn: 5,
		IdleTimeout: time.Minute,
	}
}

// NewPool membangun pgxpool dari DATABASE_URL dan menerapkan batas pool.
// MaxIdleConn tidak diterapkan karena pgxpool v5 tidak punya knob maksimum
// idle (hanya MinIdleConns); koneksi idle dibatasi via MaxConnIdleTime.
// Error tidak memuat password atau URL lengkap.
func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("DATABASE_URL kosong: pastikan .env terisi sebelum membuat pool")
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("gagal parse DATABASE_URL %s: %w", maskPassword(cfg.URL), err)
	}
	if cfg.MaxOpenCon > 0 {
		poolCfg.MaxConns = int32(cfg.MaxOpenCon)
	}
	if cfg.IdleTimeout > 0 {
		poolCfg.MaxConnIdleTime = cfg.IdleTimeout
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pgxpool: %w", err)
	}

	// Probe: pastikan database terjangkau sebelum server start.
	poolCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(poolCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("gagal ping database (pastikan PostgreSQL jalan di localhost:5432): %w", err)
	}

	return pool, nil
}

// maskPassword menyamarkan password di URL untuk log aman.
func maskPassword(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return "postgres://****"
	}
	if name := u.User.Username(); name != "" {
		if _, hasPassword := u.User.Password(); hasPassword {
			u.User = url.UserPassword(name, "****")
		}
	}
	return u.String()
}

// ClosePool menutup pool koneksi secara sinkron.
// Panggil setelah HTTP server berhenti agar request selesai dulu.
func ClosePool(pool *pgxpool.Pool) error {
	if pool == nil {
		return nil
	}
	pool.Close()
	return nil
}
