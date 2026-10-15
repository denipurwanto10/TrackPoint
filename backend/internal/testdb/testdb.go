// Package testdb adalah guard: integration test hanya boleh jalan pada DB pengujian.
package testdb

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Require abort lebih dulu bila DATABASE_URL bukan database pengujian.
// Ini mencegah test menghapus/memigrasikan database pengembangan.
func Require(expectedName string) {
	raw := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if raw == "" {
		abort("DATABASE_URL kosong — integration test dibatalkan (tidak menyentuh database apa pun)")
	}
	name := Name(raw)
	if name == "" {
		abort(fmt.Sprintf("tidak bisa membaca nama database dari DATABASE_URL: %q", raw))
	}
	if !strings.EqualFold(name, expectedName) {
		abort(fmt.Sprintf("DATABASE_URL menunjuk database %q, test hanya boleh memakai %q", name, expectedName))
	}
}

func Name(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

func abort(msg string) {
	panic("testdb guard: " + msg)
}
