// Package config memuat konfigurasi lokal dari environment (sumber: .env / PowerShell).
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL     string
	DBName          string
	DBTestName      string
	JWTAccessSecret string
	JWTAccessTTL    time.Duration
	RefreshTTL      time.Duration
	Port            string
	CORSOrigin      string
	CookieSecure    bool
}

func Load() (Config, error) {
	c := Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		DBName:          env("DB_NAME", "trackpoint"),
		DBTestName:      env("DB_TEST_NAME", "trackpoint_test"),
		JWTAccessSecret: os.Getenv("JWT_ACCESS_SECRET"),
		Port:            env("PORT", "8080"),
		CORSOrigin:      env("CORS_ORIGIN", "http://localhost:5173"),
	}

	var err error
	if c.JWTAccessTTL, err = durationMinutes("JWT_ACCESS_TTL_MIN", 15); err != nil {
		return c, err
	}
	if c.RefreshTTL, err = durationDays("REFRESH_TTL_DAYS", 30); err != nil {
		return c, err
	}
	if c.CookieSecure, err = strconv.ParseBool(env("COOKIE_SECURE", "false")); err != nil {
		return c, errors.New("COOKIE_SECURE harus true/false")
	}
	return c, c.validate()
}

func (c Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL wajib diisi")
	}
	if len(c.JWTAccessSecret) < 32 {
		return errors.New("JWT_ACCESS_SECRET wajib diisi dan minimal 32 karakter")
	}
	return nil
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func durationMinutes(key string, fallback int) (time.Duration, error) {
	n, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s harus bilangan bulat positif", key)
	}
	return time.Duration(n) * time.Minute, nil
}

func durationDays(key string, fallback int) (time.Duration, error) {
	n, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("%s harus bilangan bulat positif", key)
	}
	return time.Duration(n) * 24 * time.Hour, nil
}
