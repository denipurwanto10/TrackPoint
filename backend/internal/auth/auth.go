// Package auth: hashing password (bcrypt), JWT access token, refresh token opaque.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// dummyHash dipakai agar respons login untuk email tak dikenal memakan waktu
// sebanding (mengurangi pembeda timing), dengan pesan error yang tetap generik.
var dummyHash []byte

func init() {
	h, err := bcrypt.GenerateFromPassword([]byte("trackpoint-dummy-password"), bcrypt.DefaultCost)
	if err == nil {
		dummyHash = h
	}
}

// Claims access token: sub = ID user, role untuk RBAC tanpa query tambahan.
type Claims struct {
	UserID string `json:"-"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// HashPassword dengan bcrypt cost default.
func HashPassword(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// CheckPassword true bila cocok; false untuk hash salah/kosong (tanpa error detail).
func CheckPassword(hash, password string) bool {
	if hash == "" {
		if dummyHash != nil {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		}
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// NewAccessToken: JWT HS256 berumur ttl.
func NewAccessToken(secret, userID, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	})
	return tok.SignedString([]byte(secret))
}

// ParseAccessToken validasi signature + expiry; bedakan expired vs invalid.
func ParseAccessToken(secret, token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("metode token tak dikenal")
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("expired")
		}
		return nil, errors.New("invalid")
	}
	if claims.Subject == "" {
		return nil, errors.New("invalid")
	}
	claims.UserID = claims.Subject
	return claims, nil
}

// NewRefreshToken: string opaque acak 32 byte, base64 URL-safe (43 char).
func NewRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashRefreshToken: SHA-256 hex; yang disimpan di DB hanya ini, bukan token asli.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
