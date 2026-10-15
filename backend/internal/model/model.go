// Package model: struct domain + konstanta role/status + validasi murni (tanpa DB).
package model

import (
	"regexp"
	"strconv"
	"time"
)

const (
	RoleAdmin    = "ADMIN"
	RoleTeknisi  = "TEKNISI"
	RoleKaryawan = "KARYAWAN"
)

const (
	AssetAvailable   = "AVAILABLE"
	AssetAssigned    = "ASSIGNED"
	AssetMaintenance = "IN_MAINTENANCE"
	AssetInactive    = "INACTIVE"
)

const (
	MntReported   = "REPORTED"
	MntInProgress = "IN_PROGRESS"
	MntDone       = "DONE"
	MntCancelled  = "CANCELLED"
)

const (
	ReturnAvailable = "AVAILABLE"
	ReturnAssigned  = "ASSIGNED"
	ReturnInactive  = "INACTIVE"
)

func ValidRole(r string) bool {
	return r == RoleAdmin || r == RoleTeknisi || r == RoleKaryawan
}

func ValidAssetStatus(s string) bool {
	return s == AssetAvailable || s == AssetAssigned || s == AssetMaintenance || s == AssetInactive
}

var uuidRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID validasi format ringan sebelum nilai dikirim ke PostgreSQL.
func IsUUID(s string) bool { return uuidRe.MatchString(s) }

// maintenanceNext: transisi status tiket yang diizinkan.
var maintenanceNext = map[string][]string{
	MntReported:   {MntInProgress, MntCancelled},
	MntInProgress: {MntDone, MntCancelled},
	MntDone:       {},
	MntCancelled:  {},
}

// CanTransitionMaintenance true bila pindah status tiket diizinkan.
func CanTransitionMaintenance(from, to string) bool {
	for _, n := range maintenanceNext[from] {
		if n == to {
			return true
		}
	}
	return false
}

// ParsePagination: default page=1 per=20, per maks 100.
func ParsePagination(pageStr, perStr string) (page, per int) {
	page, per = 1, 20
	if n, err := strconv.Atoi(pageStr); err == nil && n > 0 {
		page = n
	}
	if n, err := strconv.Atoi(perStr); err == nil && n > 0 && n <= 100 {
		per = n
	}
	return page, per
}

// PageMeta hitung total halaman untuk respons list.
type PageMeta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

func NewPageMeta(page, per int, total int64) PageMeta {
	tp := 0
	if total > 0 {
		tp = int((total + int64(per) - 1) / int64(per))
	}
	return PageMeta{Page: page, PerPage: per, Total: total, TotalPages: tp}
}

type User struct {
	ID        string    `json:"id"`
	Nama      string    `json:"nama"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID        string    `json:"id"`
	Nama      string    `json:"nama"`
	Deskripsi *string   `json:"deskripsi,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Location struct {
	ID        string    `json:"id"`
	Nama      string    `json:"nama"`
	Gedung    *string   `json:"gedung,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Asset struct {
	ID           string     `json:"id"`
	Kode         string     `json:"kode"`
	Nama         string     `json:"nama"`
	SerialNumber *string    `json:"serial_number,omitempty"`
	CategoryID   string     `json:"category_id"`
	LocationID   string     `json:"location_id"`
	HolderID     *string    `json:"holder_id,omitempty"`
	Status       string     `json:"status"`
	PurchaseDate *time.Time `json:"purchase_date,omitempty"`
	Catatan      *string    `json:"catatan,omitempty"`
	IsActive     bool       `json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	// Join tampilan:
	CategoryNama string  `json:"category_nama"`
	LocationNama string  `json:"location_nama"`
	HolderNama   *string `json:"holder_nama,omitempty"`
}

type AssetMovement struct {
	ID               string    `json:"id"`
	AssetID          string    `json:"asset_id"`
	FromLocationID   *string   `json:"from_location_id,omitempty"`
	ToLocationID     string    `json:"to_location_id"`
	FromHolderID     *string   `json:"from_holder_id,omitempty"`
	ToHolderID       *string   `json:"to_holder_id,omitempty"`
	Catatan          *string   `json:"catatan,omitempty"`
	IsCorrection     bool      `json:"is_correction"`
	CorrectionReason *string   `json:"correction_reason,omitempty"`
	CreatedBy        string    `json:"created_by"`
	CreatedAt        time.Time `json:"created_at"`
	// Join tampilan:
	AssetKode        string  `json:"asset_kode"`
	AssetNama        string  `json:"asset_nama"`
	CreatedByNama    string  `json:"created_by_nama"`
	FromLocationNama *string `json:"from_location_nama,omitempty"`
	ToLocationNama   string  `json:"to_location_nama"`
}

type Maintenance struct {
	ID           string     `json:"id"`
	AssetID      string     `json:"asset_id"`
	AssetCode    string     `json:"asset_code"`
	Status       string     `json:"status"`
	Keluhan      string     `json:"keluhan"`
	Tindakan     *string    `json:"tindakan,omitempty"`
	Biaya        *float64   `json:"biaya,omitempty"`
	ReturnTo     *string    `json:"return_to,omitempty"`
	ReportedBy   string     `json:"reported_by"`
	TechnicianID *string    `json:"technician_id,omitempty"`
	CancelReason *string    `json:"cancel_reason,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	// Join tampilan:
	AssetNama      string  `json:"asset_nama"`
	ReporterNama   string  `json:"reporter_nama"`
	TechnicianNama *string `json:"technician_nama,omitempty"`
}

type RefreshSession struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	UserAgent *string    `json:"user_agent,omitempty"`
	IPAddress *string    `json:"ip_address,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UserEmail string     `json:"user_email,omitempty"`
}

// ContextUser disimpan middleware auth di gin context.
type ContextUser struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

type Dashboard struct {
	TotalAssets   int64           `json:"total_aset"`
	Available     int64           `json:"tersedia"`
	Assigned      int64           `json:"dipakai"`
	InMaintenance int64           `json:"maintenance"`
	Inactive      int64           `json:"nonaktif"`
	OpenTickets   int64           `json:"tiket_terbuka"`
	MyAssets      int64           `json:"aset_saya"`
	MyOpenTickets int64           `json:"tiket_saya_terbuka"`
	RecentMoves   []AssetMovement `json:"mutasi_terbaru"`
	ActionTickets []Maintenance   `json:"tiket_perlu_ditangani"`
}
