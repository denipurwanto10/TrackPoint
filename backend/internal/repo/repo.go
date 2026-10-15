// Package repo: semua query PostgreSQL. Menerima DB (pool/tx) agar service
// bisa menjalankan operasi multi-tabel dalam satu transaksi.
package repo

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"trackpoint/internal/model"
)

// DB dipenuhi *pgxpool.Pool dan pgx.Tx.
type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

func strPtr(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

// ---- Users ----

type UserRow struct {
	model.User
	PasswordHash string
}

func scanUserRow(r pgx.Row) (*UserRow, error) {
	var u UserRow
	err := r.Scan(&u.ID, &u.Nama, &u.Email, &u.PasswordHash, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func FindUserByEmail(ctx context.Context, db DB, email string) (*UserRow, error) {
	return scanUserRow(db.QueryRow(ctx,
		`SELECT id, nama, email, password_hash, role, is_active, created_at, updated_at
		 FROM users WHERE lower(email) = lower($1)`, email))
}

func FindUserByID(ctx context.Context, db DB, id string) (*UserRow, error) {
	return scanUserRow(db.QueryRow(ctx,
		`SELECT id, nama, email, password_hash, role, is_active, created_at, updated_at
		 FROM users WHERE id = $1`, id))
}

func ListUsers(ctx context.Context, db DB, q, role string, active *bool, page, per int) ([]model.User, int64, error) {
	var conds []string
	var args []any
	if q != "" {
		args = append(args, "%"+strings.ToLower(q)+"%")
		conds = append(conds, `(lower(nama) LIKE $`+itoa2(len(args))+` OR lower(email) LIKE $`+itoa2(len(args))+`)`)
	}
	if role != "" {
		args = append(args, role)
		conds = append(conds, `role = $`+itoa2(len(args)))
	}
	if active != nil {
		args = append(args, *active)
		conds = append(conds, `is_active = $`+itoa2(len(args)))
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	var total int64
	if err := db.QueryRow(ctx, `SELECT count(*) FROM users `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, per, (page-1)*per)
	rows, err := db.Query(ctx,
		`SELECT id, nama, email, role, is_active, created_at, updated_at FROM users `+where+
			` ORDER BY created_at DESC LIMIT $`+itoa2(len(args)-1)+` OFFSET $`+itoa2(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Nama, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	if out == nil {
		out = []model.User{}
	}
	return out, total, rows.Err()
}

func CreateUser(ctx context.Context, db DB, nama, email, hash, role string) (*model.User, error) {
	var u model.User
	err := db.QueryRow(ctx,
		`INSERT INTO users (nama, email, password_hash, role) VALUES ($1, $2, $3, $4)
		 RETURNING id, nama, email, role, is_active, created_at, updated_at`,
		nama, email, hash, role).Scan(&u.ID, &u.Nama, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func UpdateUser(ctx context.Context, db DB, id, nama, role string, active bool) (*model.User, error) {
	var u model.User
	err := db.QueryRow(ctx,
		`UPDATE users SET nama=$2, role=$3, is_active=$4, updated_at=now() WHERE id=$1
		 RETURNING id, nama, email, role, is_active, created_at, updated_at`,
		id, nama, role, active).Scan(&u.ID, &u.Nama, &u.Email, &u.Role, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func SetUserPassword(ctx context.Context, db DB, id, hash string) error {
	_, err := db.Exec(ctx, `UPDATE users SET password_hash=$2, updated_at=now() WHERE id=$1`, id, hash)
	return err
}

func CountAdmins(ctx context.Context, db DB) (int64, error) {
	var n int64
	err := db.QueryRow(ctx, `SELECT count(*) FROM users WHERE role='ADMIN' AND is_active`).Scan(&n)
	return n, err
}

// ---- Categories ----

func ListCategories(ctx context.Context, db DB, activeOnly bool) ([]model.Category, error) {
	sql := `SELECT id, nama, deskripsi, is_active, created_at, updated_at FROM categories`
	if activeOnly {
		sql += ` WHERE is_active`
	}
	sql += ` ORDER BY nama`
	rows, err := db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Nama, &c.Deskripsi, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if out == nil {
		out = []model.Category{}
	}
	return out, rows.Err()
}

func CreateCategory(ctx context.Context, db DB, nama string, deskripsi *string) (*model.Category, error) {
	var c model.Category
	err := db.QueryRow(ctx,
		`INSERT INTO categories (nama, deskripsi) VALUES ($1, $2)
		 RETURNING id, nama, deskripsi, is_active, created_at, updated_at`,
		nama, strPtr(deskripsi)).Scan(&c.ID, &c.Nama, &c.Deskripsi, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func UpdateCategory(ctx context.Context, db DB, id, nama string, deskripsi *string, active bool) (*model.Category, error) {
	var c model.Category
	err := db.QueryRow(ctx,
		`UPDATE categories SET nama=$2, deskripsi=$3, is_active=$4, updated_at=now() WHERE id=$1
		 RETURNING id, nama, deskripsi, is_active, created_at, updated_at`,
		id, nama, strPtr(deskripsi), active).Scan(&c.ID, &c.Nama, &c.Deskripsi, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func CountAssetsByCategory(ctx context.Context, db DB, id string) (int64, error) {
	var n int64
	err := db.QueryRow(ctx, `SELECT count(*) FROM assets WHERE category_id=$1 AND is_active`, id).Scan(&n)
	return n, err
}

func DeleteCategory(ctx context.Context, db DB, id string) error {
	_, err := db.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	return err
}

// ---- Locations ----

func ListLocations(ctx context.Context, db DB, activeOnly bool) ([]model.Location, error) {
	sql := `SELECT id, nama, gedung, is_active, created_at, updated_at FROM locations`
	if activeOnly {
		sql += ` WHERE is_active`
	}
	sql += ` ORDER BY nama`
	rows, err := db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Location
	for rows.Next() {
		var l model.Location
		if err := rows.Scan(&l.ID, &l.Nama, &l.Gedung, &l.IsActive, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	if out == nil {
		out = []model.Location{}
	}
	return out, rows.Err()
}

func CreateLocation(ctx context.Context, db DB, nama string, gedung *string) (*model.Location, error) {
	var l model.Location
	err := db.QueryRow(ctx,
		`INSERT INTO locations (nama, gedung) VALUES ($1, $2)
		 RETURNING id, nama, gedung, is_active, created_at, updated_at`,
		nama, strPtr(gedung)).Scan(&l.ID, &l.Nama, &l.Gedung, &l.IsActive, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func UpdateLocation(ctx context.Context, db DB, id, nama string, gedung *string, active bool) (*model.Location, error) {
	var l model.Location
	err := db.QueryRow(ctx,
		`UPDATE locations SET nama=$2, gedung=$3, is_active=$4, updated_at=now() WHERE id=$1
		 RETURNING id, nama, gedung, is_active, created_at, updated_at`,
		id, nama, strPtr(gedung), active).Scan(&l.ID, &l.Nama, &l.Gedung, &l.IsActive, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &l, nil
}

func CountAssetsByLocation(ctx context.Context, db DB, id string) (int64, error) {
	var n int64
	err := db.QueryRow(ctx, `SELECT count(*) FROM assets WHERE location_id=$1 AND is_active`, id).Scan(&n)
	return n, err
}

func DeleteLocation(ctx context.Context, db DB, id string) error {
	_, err := db.Exec(ctx, `DELETE FROM locations WHERE id=$1`, id)
	return err
}
