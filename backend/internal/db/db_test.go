package db

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxOpenCon != 25 {
		t.Errorf("MaxOpenCon = %d, want 25", cfg.MaxOpenCon)
	}
	if cfg.IdleTimeout != time.Minute {
		t.Errorf("IdleTimeout = %v, want 1m", cfg.IdleTimeout)
	}
}

func TestMaskPasswordHidesSecret(t *testing.T) {
	masked := maskPassword("postgres://trackpoint:rahasia123@localhost:5432/trackpoint?sslmode=disable")
	if strings.Contains(masked, "rahasia123") {
		t.Errorf("password bocor di hasil mask: %q", masked)
	}
	if !strings.Contains(masked, "trackpoint") {
		t.Errorf("info non-secret hilang: %q", masked)
	}
}

func TestNewPoolRejectsEmptyURL(t *testing.T) {
	if _, err := NewPool(t.Context(), Config{}); err == nil {
		t.Error("NewPool harus error saat URL kosong")
	}
}
