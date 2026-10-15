// Repo aset, mutasi, maintenance, sesi refresh, dashboard.
package repo

import (
	"context"
	"strings"

	"trackpoint/internal/model"
)

const assetCols = `a.id, a.kode, a.nama, a.serial_number, a.category_id, a.location_id,
	a.holder_id, a.status, a.purchase_date, a.catatan, a.is_active, a.created_at, a.updated_at,
	c.nama AS category_nama, l.nama AS location_nama, h.nama AS holder_nama`
const assetJoins = `FROM assets a
	JOIN categories c ON c.id = a.category_id
	JOIN locations l ON l.id = a.location_id
	LEFT JOIN users h ON h.id = a.holder_id`

func scanAsset(scan func(dest ...any) error) (model.Asset, error) {
	var a model.Asset
	err := scan(&a.ID, &a.Kode, &a.Nama, &a.SerialNumber, &a.CategoryID, &a.LocationID,
		&a.HolderID, &a.Status, &a.PurchaseDate, &a.Catatan, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
		&a.CategoryNama, &a.LocationNama, &a.HolderNama)
	return a, err
}

func ListAssets(ctx context.Context, db DB, q, status, categoryID, locationID, holderID string, page, per int) ([]model.Asset, int64, error) {
	var conds []string
	var args []any
	push := func(cond string, v any) {
		args = append(args, v)
		conds = append(conds, cond+"$"+itoa2(len(args)))
	}
	if q != "" {
		args = append(args, "%"+strings.ToLower(q)+"%")
		n := itoa2(len(args))
		conds = append(conds, `(lower(a.nama) LIKE $`+n+` OR lower(a.kode) LIKE $`+n+` OR lower(coalesce(a.serial_number,'')) LIKE $`+n+`)`)
	}
	if model.ValidAssetStatus(status) {
		push(`a.status = `, status)
	}
	if categoryID != "" {
		push(`a.category_id = `, categoryID)
	}
	if locationID != "" {
		push(`a.location_id = `, locationID)
	}
	if holderID != "" {
		push(`a.holder_id = `, holderID)
	}
	conds = append(conds, `a.is_active`)
	where := "WHERE " + strings.Join(conds, " AND ")
	var total int64
	if err := db.QueryRow(ctx, `SELECT count(*) `+assetJoins+` `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, per, (page-1)*per)
	rows, err := db.Query(ctx,
		`SELECT `+assetCols+` `+assetJoins+` `+where+
			` ORDER BY a.created_at DESC LIMIT $`+itoa2(len(args)-1)+` OFFSET $`+itoa2(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.Asset{}
	for rows.Next() {
		a, err := scanAsset(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func GetAsset(ctx context.Context, db DB, id string) (*model.Asset, error) {
	a, err := scanAsset(db.QueryRow(ctx,
		`SELECT `+assetCols+` `+assetJoins+` WHERE a.id=$1`, id).Scan)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// GetAssetForUpdate mengunci baris aset (FOR UPDATE) untuk mutasi konkuren aman.
func GetAssetForUpdate(ctx context.Context, db DB, id string) (*model.Asset, error) {
	a, err := scanAsset(db.QueryRow(ctx,
		`SELECT `+assetCols+` `+assetJoins+` WHERE a.id=$1 FOR UPDATE OF a`, id).Scan)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// NextAssetCode dari sequence Postgres: TRK-000001, TRK-000002, ...
func NextAssetCode(ctx context.Context, db DB) (string, error) {
	var n int64
	if err := db.QueryRow(ctx, `SELECT nextval('asset_code_seq')`).Scan(&n); err != nil {
		return "", err
	}
	return "TRK-" + pad6(n), nil
}

func pad6(n int64) string {
	s := itoa64(n)
	for len(s) < 6 {
		s = "0" + s
	}
	return s
}

func itoa64(n int64) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// CreateAsset hanya mengisi kolom non-status (status awal selalu AVAILABLE).
func CreateAsset(ctx context.Context, db DB, kode, nama string, serial, categoryID, locationID string, purchaseDate *string, catatan *string) (*model.Asset, error) {
	a, err := scanAsset(db.QueryRow(ctx,
		`INSERT INTO assets (kode, nama, serial_number, category_id, location_id, purchase_date, catatan)
		 VALUES ($1,$2,nullif($3,''),$4,$5, nullif($6,'')::date, $7)
		 RETURNING `+assetColsReturn(),
		kode, nama, serial, categoryID, locationID, purchaseDateStr(purchaseDate), strPtr(catatan),
	).Scan)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func assetColsReturn() string {
	return `id, kode, nama, serial_number, category_id, location_id, holder_id, status,
		purchase_date, catatan, is_active, created_at, updated_at,
		''::text, ''::text, NULL::text`
}

func purchaseDateStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// UpdateAssetMeta hanya nama/serial/kategori/catatan/purchase_date.
// Status, lokasi, holder TIDAK boleh lewat endpoint edit umum.
func UpdateAssetMeta(ctx context.Context, db DB, id, nama, serial, categoryID string, purchaseDate *string, catatan *string) (*model.Asset, error) {
	a, err := scanAsset(db.QueryRow(ctx,
		`UPDATE assets SET nama=$2, serial_number=nullif($3,''), category_id=$4,
			purchase_date=nullif($5,'')::date, catatan=$6, updated_at=now() WHERE id=$1
		 RETURNING `+assetColsReturn(), id, nama, serial, categoryID, purchaseDateStr(purchaseDate), strPtr(catatan)).Scan)
	if err != nil {
		return nil, err
	}
	return GetAsset(ctx, db, a.ID)
}

// ApplyAssetMove mengubah status/lokasi/holder dalam transaksi yang sama dengan insert mutasi.
func ApplyAssetMove(ctx context.Context, db DB, id, status, locationID string, holderID *string) error {
	_, err := db.Exec(ctx,
		`UPDATE assets SET status=$2, location_id=$3, holder_id=$4, updated_at=now() WHERE id=$1`,
		id, status, locationID, holderID)
	return err
}

// CountAssetsByHolder dipakai untuk validasi hapus/nonaktif user.
func CountAssetsByHolder(ctx context.Context, db DB, holderID string) (int64, error) {
	var n int64
	err := db.QueryRow(ctx, `SELECT count(*) FROM assets WHERE holder_id=$1 AND is_active`, holderID).Scan(&n)
	return n, err
}

// ---- Movements (append-only) ----

func InsertMovement(ctx context.Context, db DB, m model.AssetMovement) (*model.AssetMovement, error) {
	var out model.AssetMovement
	err := db.QueryRow(ctx,
		`INSERT INTO asset_movements
		 (asset_id, from_location_id, to_location_id, from_holder_id, to_holder_id, catatan, is_correction, correction_reason, created_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, asset_id, from_location_id, to_location_id, from_holder_id, to_holder_id,
			catatan, is_correction, correction_reason, created_by, created_at`,
		m.AssetID, m.FromLocationID, m.ToLocationID, m.FromHolderID, m.ToHolderID,
		strPtr(m.Catatan), m.IsCorrection, strPtr(m.CorrectionReason), m.CreatedBy,
	).Scan(&out.ID, &out.AssetID, &out.FromLocationID, &out.ToLocationID, &out.FromHolderID,
		&out.ToHolderID, &out.Catatan, &out.IsCorrection, &out.CorrectionReason, &out.CreatedBy, &out.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func ListMovements(ctx context.Context, db DB, assetID string, page, per int) ([]model.AssetMovement, int64, error) {
	var total int64
	if err := db.QueryRow(ctx, `SELECT count(*) FROM asset_movements WHERE asset_id=$1`, assetID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := db.Query(ctx,
		`SELECT m.id, m.asset_id, m.from_location_id, m.to_location_id, m.from_holder_id, m.to_holder_id,
			m.catatan, m.is_correction, m.correction_reason, m.created_by, m.created_at,
			a.kode, a.nama, u.nama, fl.nama, tl.nama
		 FROM asset_movements m
		 JOIN assets a ON a.id = m.asset_id
		 JOIN users u ON u.id = m.created_by
		 LEFT JOIN locations fl ON fl.id = m.from_location_id
		 JOIN locations tl ON tl.id = m.to_location_id
		 WHERE m.asset_id=$1 ORDER BY m.created_at DESC LIMIT $2 OFFSET $3`, assetID, per, (page-1)*per)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.AssetMovement{}
	for rows.Next() {
		var m model.AssetMovement
		if err := rows.Scan(&m.ID, &m.AssetID, &m.FromLocationID, &m.ToLocationID, &m.FromHolderID,
			&m.ToHolderID, &m.Catatan, &m.IsCorrection, &m.CorrectionReason, &m.CreatedBy, &m.CreatedAt,
			&m.AssetKode, &m.AssetNama, &m.CreatedByNama, &m.FromLocationNama, &m.ToLocationNama); err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func RecentMovements(ctx context.Context, db DB, limit int, holderID string) ([]model.AssetMovement, error) {
	args := []any{}
	cond := ""
	if holderID != "" {
		args = append(args, holderID)
		cond = `WHERE (a.holder_id = $1 OR m.to_holder_id = $1 OR m.from_holder_id = $1)`
	}
	args = append(args, limit)
	rows, err := db.Query(ctx,
		`SELECT m.id, m.asset_id, m.from_location_id, m.to_location_id, m.from_holder_id, m.to_holder_id,
			m.catatan, m.is_correction, m.correction_reason, m.created_by, m.created_at,
			a.kode, a.nama, u.nama, fl.nama, tl.nama
		 FROM asset_movements m
		 JOIN assets a ON a.id = m.asset_id
		 JOIN users u ON u.id = m.created_by
		 LEFT JOIN locations fl ON fl.id = m.from_location_id
		 JOIN locations tl ON tl.id = m.to_location_id
		 `+cond+` ORDER BY m.created_at DESC LIMIT $`+itoa2(len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.AssetMovement{}
	for rows.Next() {
		var m model.AssetMovement
		if err := rows.Scan(&m.ID, &m.AssetID, &m.FromLocationID, &m.ToLocationID, &m.FromHolderID,
			&m.ToHolderID, &m.Catatan, &m.IsCorrection, &m.CorrectionReason, &m.CreatedBy, &m.CreatedAt,
			&m.AssetKode, &m.AssetNama, &m.CreatedByNama, &m.FromLocationNama, &m.ToLocationNama); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---- Maintenance ----

func scanMaintenance(scan func(dest ...any) error) (model.Maintenance, error) {
	var m model.Maintenance
	err := scan(&m.ID, &m.AssetID, &m.AssetCode, &m.Status, &m.Keluhan, &m.Tindakan, &m.Biaya,
		&m.ReturnTo, &m.ReportedBy, &m.TechnicianID, &m.CancelReason, &m.StartedAt, &m.CompletedAt,
		&m.CreatedAt, &m.UpdatedAt, &m.AssetNama, &m.ReporterNama, &m.TechnicianNama)
	return m, err
}

const mntSelect = `SELECT m.id, m.asset_id, m.asset_code, m.status, m.keluhan, m.tindakan, m.biaya,
		m.return_to, m.reported_by, m.technician_id, m.cancel_reason, m.started_at, m.completed_at,
		m.created_at, m.updated_at, a.nama, r.nama, t.nama
	FROM maintenance_records m
	JOIN assets a ON a.id = m.asset_id
	JOIN users r ON r.id = m.reported_by
	LEFT JOIN users t ON t.id = m.technician_id`

func ListMaintenance(ctx context.Context, db DB, status, mine string, page, per int) ([]model.Maintenance, int64, error) {
	var conds []string
	var args []any
	if status == model.MntReported || status == model.MntInProgress || status == model.MntDone || status == model.MntCancelled {
		args = append(args, status)
		conds = append(conds, `m.status = $`+itoa2(len(args)))
	}
	if mine != "" {
		args = append(args, mine)
		conds = append(conds, `(m.reported_by = $`+itoa2(len(args))+` OR m.technician_id = $`+itoa2(len(args))+`)`)
	}
	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}
	var total int64
	if err := db.QueryRow(ctx, `SELECT count(*) FROM maintenance_records m `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, per, (page-1)*per)
	rows, err := db.Query(ctx, mntSelect+` `+where+
		` ORDER BY m.created_at DESC LIMIT $`+itoa2(len(args)-1)+` OFFSET $`+itoa2(len(args)), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []model.Maintenance{}
	for rows.Next() {
		m, err := scanMaintenance(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, m)
	}
	return out, total, rows.Err()
}

func GetMaintenance(ctx context.Context, db DB, id string) (*model.Maintenance, error) {
	m, err := scanMaintenance(db.QueryRow(ctx, mntSelect+` WHERE m.id=$1`, id).Scan)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func GetMaintenanceForUpdate(ctx context.Context, db DB, id string) (*model.Maintenance, error) {
	m, err := scanMaintenance(db.QueryRow(ctx, mntSelect+` WHERE m.id=$1 FOR UPDATE OF m`, id).Scan)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func CreateMaintenance(ctx context.Context, db DB, assetID, assetCode, keluhan, reportedBy string) (*model.Maintenance, error) {
	var id string
	err := db.QueryRow(ctx,
		`INSERT INTO maintenance_records (asset_id, asset_code, keluhan, reported_by)
		 VALUES ($1,$2,$3,$4) RETURNING id`, assetID, assetCode, keluhan, reportedBy).Scan(&id)
	if err != nil {
		return nil, err
	}
	return GetMaintenance(ctx, db, id)
}

// UpdateMaintenanceStatus transisi tiket; kolom pendukung diisi sesuai status tujuan.
func UpdateMaintenanceStatus(ctx context.Context, db DB, id, to, tindakan string, biaya *float64, returnTo, cancelReason, technicianID string) (*model.Maintenance, error) {
	var set string
	var args []any
	args = append(args, id)
	switch to {
	case model.MntInProgress:
		set = `status='IN_PROGRESS', technician_id=$2, started_at=coalesce(started_at,now())`
		args = append(args, nullStr(technicianID))
	case model.MntDone:
		set = `status='DONE', tindakan=$2, biaya=$3, return_to=$4, completed_at=now()`
		args = append(args, tindakan, biaya, nullStr(returnTo))
	case model.MntCancelled:
		set = `status='CANCELLED', cancel_reason=$2`
		args = append(args, cancelReason)
	default:
		set = `status=$2`
		args = append(args, to)
	}
	var newID string
	err := db.QueryRow(ctx,
		`UPDATE maintenance_records SET `+set+`, updated_at=now() WHERE id=$1 RETURNING id`, args...).Scan(&newID)
	if err != nil {
		return nil, err
	}
	return GetMaintenance(ctx, db, newID)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// OpenTicketsForAction: tiket REPORTED/IN_PROGRESS untuk dashboard.
// mineOnly=true (karyawan): hanya tiket yang dilaporkan user tsb.
func OpenTicketsForAction(ctx context.Context, db DB, userID string, mineOnly bool, limit int) ([]model.Maintenance, error) {
	var rowsArgs []any
	where := `WHERE m.status IN ('REPORTED','IN_PROGRESS')`
	if mineOnly {
		rowsArgs = append(rowsArgs, userID)
		where += ` AND m.reported_by = $1`
	}
	rowsArgs = append(rowsArgs, limit)
	rows, err := db.Query(ctx, mntSelect+` `+where+` ORDER BY m.created_at DESC LIMIT $`+itoa2(len(rowsArgs)), rowsArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.Maintenance{}
	for rows.Next() {
		m, err := scanMaintenance(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ---- Refresh sessions ----

func CreateRefreshSession(ctx context.Context, db DB, userID, tokenHash string, expiresAt any, ua, ip *string) (string, error) {
	var id string
	err := db.QueryRow(ctx,
		`INSERT INTO refresh_sessions (user_id, token_hash, expires_at, user_agent, ip_address)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`, userID, tokenHash, expiresAt, ua, ip).Scan(&id)
	return id, err
}

type RefreshRow struct {
	ID         string
	UserID     string
	RevokedAt  *string
	ReplacedBy *string
	Expired    bool
}

// ErrRotateConflict: token sudah diputar request lain (refresh konkuren).
var ErrRotateConflict = errSentinelRepo("rotate-conflict")

type errSentinelRepo string

func (s errSentinelRepo) Error() string { return string(s) }

func FindRefreshSession(ctx context.Context, db DB, tokenHash string) (*RefreshRow, error) {
	var r RefreshRow
	err := db.QueryRow(ctx,
		`SELECT id, user_id, revoked_at::text, replaced_by::text, (expires_at <= now()) AS expired
		 FROM refresh_sessions WHERE token_hash=$1`, tokenHash).
		Scan(&r.ID, &r.UserID, &r.RevokedAt, &r.ReplacedBy, &r.Expired)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// RotateRefreshSession: cabut sesi lama (set replaced_by) + buat sesi baru, atomik via tx.
// UPDATE bersyarat revoked_at IS NULL: refresh konkuren kedua gagal (0 baris)
// sehingga tidak ada dua sesi pengganti dari satu token.
func RotateRefreshSession(ctx context.Context, db DB, oldID, userID, newHash string, expiresAt any, ua, ip *string) (string, error) {
	var newID string
	err := db.QueryRow(ctx,
		`INSERT INTO refresh_sessions (user_id, token_hash, expires_at, user_agent, ip_address)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`, userID, newHash, expiresAt, ua, ip).Scan(&newID)
	if err != nil {
		return "", err
	}
	tag, err := db.Exec(ctx,
		`UPDATE refresh_sessions SET revoked_at=now(), replaced_by=$2 WHERE id=$1 AND revoked_at IS NULL`, oldID, newID)
	if err != nil {
		return "", err
	}
	if tag.RowsAffected() != 1 {
		return "", ErrRotateConflict
	}
	return newID, nil
}

// RevokeSession cabut satu sesi (logout / cabut oleh admin).
func RevokeSession(ctx context.Context, db DB, id string) error {
	_, err := db.Exec(ctx, `UPDATE refresh_sessions SET revoked_at=now() WHERE id=$1`, id)
	return err
}

// RevokeAllUserSessions cabut semua sesi aktif user (dipakai saat reuse terdeteksi / nonaktif).
func RevokeAllUserSessions(ctx context.Context, db DB, userID string) error {
	_, err := db.Exec(ctx, `UPDATE refresh_sessions SET revoked_at=now() WHERE user_id=$1 AND revoked_at IS NULL`, userID)
	return err
}

func ListActiveSessions(ctx context.Context, db DB, userID string) ([]model.RefreshSession, error) {
	rows, err := db.Query(ctx,
		`SELECT s.id, s.user_id, s.expires_at, s.revoked_at, s.user_agent, s.ip_address, s.created_at, u.email
		 FROM refresh_sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.user_id=$1 AND s.revoked_at IS NULL AND s.expires_at > now()
		 ORDER BY s.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.RefreshSession{}
	for rows.Next() {
		var s model.RefreshSession
		if err := rows.Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.RevokedAt, &s.UserAgent, &s.IPAddress, &s.CreatedAt, &s.UserEmail); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ---- Dashboard ----

func AssetStatusCounts(ctx context.Context, db DB) (map[string]int64, error) {
	rows, err := db.Query(ctx, `SELECT status, count(*) FROM assets WHERE is_active GROUP BY status`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := map[string]int64{}
	for rows.Next() {
		var s string
		var n int64
		if err := rows.Scan(&s, &n); err != nil {
			return nil, err
		}
		m[s] = n
	}
	return m, rows.Err()
}

func CountOpenTickets(ctx context.Context, db DB) (int64, error) {
	var n int64
	err := db.QueryRow(ctx, `SELECT count(*) FROM maintenance_records WHERE status IN ('REPORTED','IN_PROGRESS')`).Scan(&n)
	return n, err
}

func CountMyAssets(ctx context.Context, db DB, userID string) (int64, error) {
	var n int64
	err := db.QueryRow(ctx, `SELECT count(*) FROM assets WHERE holder_id=$1 AND is_active`, userID).Scan(&n)
	return n, err
}

func CountMyOpenTickets(ctx context.Context, db DB, userID string) (int64, error) {
	var n int64
	err := db.QueryRow(ctx,
		`SELECT count(*) FROM maintenance_records WHERE status IN ('REPORTED','IN_PROGRESS') AND reported_by=$1`, userID).Scan(&n)
	return n, err
}

// itoa2 untuk placeholder $N (N kecil, maks 2 digit cukup untuk query di sini).
func itoa2(n int) string {
	if n <= 0 {
		return "1"
	}
	if n < 10 {
		return string(rune('0' + n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
