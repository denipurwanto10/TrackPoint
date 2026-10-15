package testdb

import (
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	cases := map[string]string{
		"postgres://u:p@localhost:5432/trackpoint?sslmode=disable":      "trackpoint",
		"postgres://u:p@localhost:5432/trackpoint_test?sslmode=disable": "trackpoint_test",
		"postgres://u:p@localhost:5432/":                                "",
		"::not a url::":                                                 "",
	}
	for raw, want := range cases {
		if got := Name(raw); got != want {
			t.Errorf("Name(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestRequireRejectsDevDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/trackpoint?sslmode=disable")
	defer func() {
		msg := recover()
		if msg == nil {
			t.Fatal("harus panic saat DATABASE_URL menunjuk database pengembangan")
		}
		if s, ok := msg.(string); !ok || !strings.Contains(s, "trackpoint_test") {
			t.Fatalf("pesan panic tidak menjelaskan aturan: %v", msg)
		}
	}()
	Require("trackpoint_test")
}

func TestRequireRejectsEmptyURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	defer func() {
		if recover() == nil {
			t.Fatal("harus panic saat DATABASE_URL kosong")
		}
	}()
	Require("trackpoint_test")
}

func TestRequireAcceptsTestDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/trackpoint_test?sslmode=disable")
	Require("trackpoint_test") // tidak panic = lulus
}
