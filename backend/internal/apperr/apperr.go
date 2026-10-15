// Package apperr: error API terstruktur {code, message} + status HTTP.
// code berbahasa Inggris (mesin), message Indonesia (manusia, aman untuk klien).
package apperr

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// Error adalah error API yang aman dikirim ke klien.
type Error struct {
	Code    string
	Message string
	Status  int
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

func New(status int, code, message string) *Error {
	return &Error{Code: code, Message: message, Status: status}
}

func BadRequest(code, message string) *Error    { return New(400, code, message) }
func Unauthorized(code, message string) *Error  { return New(401, code, message) }
func Forbidden(code, message string) *Error     { return New(403, code, message) }
func NotFound(code, message string) *Error      { return New(404, code, message) }
func Conflict(code, message string) *Error      { return New(409, code, message) }
func Unprocessable(code, message string) *Error { return New(422, code, message) }
func Internal() *Error                          { return New(500, "INTERNAL_ERROR", "Terjadi kesalahan pada server.") }

// FromPG memetakan error PostgreSQL ke error API tanpa membocorkan detail internal.
func FromPG(err error, resource string) *Error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return Conflict("DUPLICATE_DATA", resource+" sudah terdaftar.")
		case "23503":
			return Conflict("DATA_TERPAKAI", "Data masih dipakai oleh data lain.")
		case "23514":
			return Unprocessable("ATURAN_DILANGGAR", "Data melanggar aturan bisnis "+resource+".")
		case "22P02", "22P05", "22021":
			return BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid.")
		}
	}
	return Internal()
}
