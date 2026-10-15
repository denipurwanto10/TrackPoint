// cmd/seed membuat akun admin awal + data master minimal untuk dev lokal.
//
// Aman: password dibuat acak tiap seed baru dan hanya ditampilkan SEKALI di
// stdout (tidak ditulis ke log/file). Idempoten: bila admin sudah ada,
// tidak membuat apa-apa dan tidak menampilkan kredensial.
//
// Jalankan manual dari direktori backend (jangan otomatis saat start):
//
//	$env:DATABASE_URL="postgres://user:pass@localhost:5432/trackpoint?sslmode=disable"
//	$env:SEED_ADMIN_EMAIL="admin@trackpoint.local"   # opsional
//	go run ./cmd/seed
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"trackpoint/internal/auth"
	"trackpoint/internal/db"
)

const adminAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"

func main() {
	log.SetFlags(0)
	if err := run(); err != nil {
		log.Fatalf("seed: %v", err)
	}
}

func run() error {
	url := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if url == "" {
		return errf("DATABASE_URL wajib diisi")
	}
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL")))
	if email == "" {
		email = "admin@trackpoint.local"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cfg := db.DefaultConfig()
	cfg.URL = url
	pool, err := db.NewPool(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.ClosePool(pool)

	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE lower(email)=lower($1))`, email).Scan(&exists); err != nil {
		return errf("cek admin: %v", maskErr(err))
	}
	if exists {
		fmt.Println("Admin sudah ada, tidak ada yang dibuat.")
		return nil
	}

	password, err := randomPassword(16)
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return errf("hash password gagal")
	}

	txCtx, txCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer txCancel()
	tx, err := pool.Begin(txCtx)
	if err != nil {
		return errf("mulai transaksi: %v", maskErr(err))
	}
	defer tx.Rollback(txCtx)

	var adminID string
	if err := tx.QueryRow(txCtx,
		`INSERT INTO users (nama, email, password_hash, role) VALUES ('Administrator', $1, $2, 'ADMIN') RETURNING id`,
		email, hash).Scan(&adminID); err != nil {
		return errf("buat admin: %v", maskErr(err))
	}
	// Data master minimal agar aplikasi langsung bisa dipakai.
	for _, c := range []string{"Laptop", "Desktop", "Monitor", "Printer", "Jaringan"} {
		_, _ = tx.Exec(txCtx, `INSERT INTO categories (nama) VALUES ($1) ON CONFLICT DO NOTHING`, c)
	}
	for _, l := range []string{"Kantor Pusat", "Gudang", "Ruang Server"} {
		_, _ = tx.Exec(txCtx, `INSERT INTO locations (nama) VALUES ($1) ON CONFLICT DO NOTHING`, l)
	}
	if err := tx.Commit(txCtx); err != nil {
		return errf("commit: %v", maskErr(err))
	}

	fmt.Println("==== AKUN ADMIN DIBUAT (ditampilkan sekali) ====")
	fmt.Printf("Email   : %s\n", email)
	fmt.Printf("Password: %s\n", password)
	fmt.Println("Segera login dan ganti password. Jangan bagikan kredensial ini.")
	return nil
}

func randomPassword(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = adminAlphabet[int(b[i])%len(adminAlphabet)]
	}
	return string(b), nil
}

type seedErr string

func (e seedErr) Error() string { return string(e) }
func errf(format string, a ...any) error {
	return seedErr(fmt.Sprintf(format, a...))
}

// maskErr memastikan error DB (yang bisa memuat potongan connection string
// dari driver) tidak bocor apa adanya ke output.
func maskErr(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if i := strings.Index(msg, "@"); i >= 0 {
		msg = "koneksi/query database gagal"
	}
	return seedErr(msg)
}
