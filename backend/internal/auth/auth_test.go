package auth

import (
	"testing"
	"time"
)

func TestHashAndCheck(t *testing.T) {
	h, err := HashPassword("rahasia123")
	if err != nil {
		t.Fatal(err)
	}
	if h == "rahasia123" {
		t.Fatal("hash sama dengan plaintext")
	}
	if !CheckPassword(h, "rahasia123") {
		t.Fatal("password benar ditolak")
	}
	if CheckPassword(h, "salah") {
		t.Fatal("password salah diterima")
	}
	// Hash salah format → false, bukan panic.
	if CheckPassword("bukan-hash", "x") {
		t.Fatal("hash invalid diterima")
	}
}

func TestAccessTokenRoundTrip(t *testing.T) {
	secret := "0123456789abcdef0123456789abcdef"
	tok, err := NewAccessToken(secret, "user-1", "ADMIN", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	cl, err := ParseAccessToken(secret, tok)
	if err != nil {
		t.Fatal(err)
	}
	if cl.UserID != "user-1" || cl.Role != "ADMIN" {
		t.Fatalf("claims salah: %+v", cl)
	}
	// Secret beda → ditolak.
	if _, err := ParseAccessToken("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", tok); err == nil {
		t.Fatal("token lolos dengan secret beda")
	}
}

func TestRefreshTokenLength(t *testing.T) {
	for range 5 {
		tok, err := NewRefreshToken()
		if err != nil {
			t.Fatal(err)
		}
		// 32 byte acak → base64url tanpa padding = 43 char.
		if len(tok) < 43 {
			t.Fatalf("refresh terlalu pendek: %d", len(tok))
		}
		h1, h2 := HashRefreshToken(tok), HashRefreshToken(tok+"x")
		if h1 == h2 || len(h1) != 64 {
			t.Fatal("hash refresh tidak stabil/valid")
		}
	}
}
