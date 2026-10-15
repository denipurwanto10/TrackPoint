// Service bisnis: aset, mutasi, maintenance, kategori, lokasi, user admin, dashboard.
package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"trackpoint/internal/apperr"
	"trackpoint/internal/auth"
	"trackpoint/internal/model"
	"trackpoint/internal/repo"
)

// ---- Assets ----

type AssetService struct {
	pool TxBeginner
	db   repo.DB
}

func NewAssetService(pool TxBeginner, db repo.DB) *AssetService { return &AssetService{pool, db} }

func (s *AssetService) List(ctx context.Context, q, status, cat, loc, holder string, page, per int, actor *model.ContextUser) ([]model.Asset, model.PageMeta, *apperr.Error) {
	// Karyawan hanya boleh melihat aset yang mereka pegang
	if actor != nil && actor.Role == model.RoleKaryawan {
		holder = actor.ID
	}
	items, total, err := repo.ListAssets(ctx, s.db, q, status, cat, loc, holder, page, per)
	if err != nil {
		return nil, model.PageMeta{}, apperr.Internal()
	}
	return items, model.NewPageMeta(page, per, total), nil
}

func (s *AssetService) Get(ctx context.Context, id string) (*model.Asset, *apperr.Error) {
	if !model.IsUUID(id) {
		return nil, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID aset tidak valid.")
	}
	a, err := repo.GetAsset(ctx, s.db, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.NotFound("ASET_TIDAK_DITEMUKAN", "Aset tidak ditemukan.")
		}
		return nil, apperr.FromPG(err, "aset")
	}
	return a, nil
}

type CreateAssetInput struct {
	Nama         string `json:"nama"`
	SerialNumber string `json:"serial_number"`
	CategoryID   string `json:"category_id"`
	LocationID   string `json:"location_id"`
	PurchaseDate string `json:"purchase_date"`
	Catatan      string `json:"catatan"`
}

func (s *AssetService) Create(ctx context.Context, in CreateAssetInput, actor model.ContextUser) (*model.Asset, *apperr.Error) {
	if strings.TrimSpace(in.Nama) == "" {
		return nil, apperr.BadRequest("NAMA_WAJIB", "Nama aset wajib diisi.")
	}
	if !model.IsUUID(in.CategoryID) || !model.IsUUID(in.LocationID) {
		return nil, apperr.BadRequest("RELASI_TIDAK_VALID", "Kategori atau lokasi tidak valid.")
	}
	// Karyawan tidak boleh membuat aset.
	if actor.Role == model.RoleKaryawan {
		return nil, apperr.Forbidden("AKSES_DITOLAK", "Anda tidak punya akses membuat aset.")
	}
	var out *model.Asset
	err := withTx(ctx, s.pool, func(tx pgx.Tx) error {
		code, cerr := repo.NextAssetCode(ctx, tx)
		if cerr != nil {
			return cerr
		}
		a, cerr := repo.CreateAsset(ctx, tx, code, strings.TrimSpace(in.Nama), in.SerialNumber, in.CategoryID, in.LocationID, optStr(in.PurchaseDate), optStr(in.Catatan))
		if cerr != nil {
			return cerr
		}
		out = a
		return nil
	})
	if err != nil {
		return nil, apperr.FromPG(err, "aset")
	}
	full, gerr := repo.GetAsset(ctx, s.db, out.ID)
	if gerr != nil {
		return nil, apperr.Internal()
	}
	return full, nil
}

type UpdateAssetInput struct {
	Nama         string `json:"nama"`
	SerialNumber string `json:"serial_number"`
	CategoryID   string `json:"category_id"`
	PurchaseDate string `json:"purchase_date"`
	Catatan      string `json:"catatan"`
}

// UpdateMeta hanya kolom non-status; status/lokasi/holder lewat endpoint mutasi.
func (s *AssetService) UpdateMeta(ctx context.Context, id string, in UpdateAssetInput, actor model.ContextUser) (*model.Asset, *apperr.Error) {
	if !model.IsUUID(id) {
		return nil, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID aset tidak valid.")
	}
	if actor.Role == model.RoleKaryawan {
		return nil, apperr.Forbidden("AKSES_DITOLAK", "Anda tidak punya akses mengubah aset.")
	}
	if strings.TrimSpace(in.Nama) == "" {
		return nil, apperr.BadRequest("NAMA_WAJIB", "Nama aset wajib diisi.")
	}
	a, err := repo.UpdateAssetMeta(ctx, s.db, id, strings.TrimSpace(in.Nama), in.SerialNumber, in.CategoryID, optStr(in.PurchaseDate), optStr(in.Catatan))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.NotFound("ASET_TIDAK_DITEMUKAN", "Aset tidak ditemukan.")
		}
		return nil, apperr.FromPG(err, "aset")
	}
	return a, nil
}

// ---- Mutasi ----

type MoveInput struct {
	ToLocationID string `json:"to_location_id"`
	ToHolderID   string `json:"to_holder_id"` // kosong = tanpa pemegang
	Catatan      string `json:"catatan"`
	IsCorrection bool   `json:"is_correction"`
	Reason       string `json:"reason"`
}

// Move: assign/pindah/koreksi dalam satu transaksi + FOR UPDATE.
func (s *AssetService) Move(ctx context.Context, assetID string, in MoveInput, actor model.ContextUser) (*model.AssetMovement, *apperr.Error) {
	if !model.IsUUID(assetID) || !model.IsUUID(in.ToLocationID) {
		return nil, apperr.BadRequest("INPUT_TIDAK_VALID", "ID aset atau lokasi tujuan tidak valid.")
	}
	if in.ToHolderID != "" && !model.IsUUID(in.ToHolderID) {
		return nil, apperr.BadRequest("HOLDER_TIDAK_VALID", "ID pemegang tidak valid.")
	}
	if actor.Role == model.RoleKaryawan {
		return nil, apperr.Forbidden("AKSES_DITOLAK", "Hanya teknisi atau admin yang dapat memutasi aset.")
	}
	if in.IsCorrection {
		if actor.Role != model.RoleAdmin {
			return nil, apperr.Forbidden("KOREKSI_ADMIN_SAJA", "Koreksi mutasi hanya oleh admin.")
		}
		if strings.TrimSpace(in.Reason) == "" {
			return nil, apperr.BadRequest("ALASAN_WAJIB", "Alasan koreksi wajib diisi.")
		}
	}
	var out *model.AssetMovement
	err := withTx(ctx, s.pool, func(tx pgx.Tx) error {
		a, gerr := repo.GetAssetForUpdate(ctx, tx, assetID)
		if gerr != nil {
			return gerr
		}
		if a.Status == model.AssetMaintenance {
			return errMaintenanceLock
		}
		var holder *string
		newStatus := model.AssetAvailable
		if in.ToHolderID != "" {
			// Pemegang harus user aktif (karyawan/teknisi/admin).
			u, uerr := repo.FindUserByID(ctx, tx, in.ToHolderID)
			if uerr != nil {
				return uerr
			}
			if !u.IsActive {
				return errHolderInactive
			}
			h := in.ToHolderID
			holder = &h
			newStatus = model.AssetAssigned
		}
		m, merr := repo.InsertMovement(ctx, tx, model.AssetMovement{
			AssetID:          assetID,
			FromLocationID:   &a.LocationID,
			ToLocationID:     in.ToLocationID,
			FromHolderID:     a.HolderID,
			ToHolderID:       holderOrNil(in.ToHolderID),
			Catatan:          optStr(in.Catatan),
			IsCorrection:     in.IsCorrection,
			CorrectionReason: optStr(in.Reason),
			CreatedBy:        actor.ID,
		})
		if merr != nil {
			return merr
		}
		if merr := repo.ApplyAssetMove(ctx, tx, assetID, newStatus, in.ToLocationID, holder); merr != nil {
			return merr
		}
		out = m
		return nil
	})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, apperr.NotFound("DATA_TIDAK_DITEMUKAN", "Aset atau pemegang tidak ditemukan.")
		case errMaintenanceLock:
			return nil, apperr.Conflict("ASET_MAINTENANCE", "Aset sedang maintenance, tidak bisa dimutasi.")
		case errHolderInactive:
			return nil, apperr.Unprocessable("HOLDER_NONAKTIF", "Pemegang tidak aktif.")
		}
		return nil, apperr.FromPG(err, "mutasi")
	}
	return out, nil
}

var (
	errMaintenanceLock = errSentinel("maintenance")
	errHolderInactive  = errSentinel("holder-inactive")
)

type sentinel string

func errSentinel(s string) error { return sentinel(s) }
func (s sentinel) Error() string { return string(s) }

func holderOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optStr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	v := strings.TrimSpace(s)
	return &v
}

// ---- Kategori & Lokasi ----

type MasterService struct {
	db repo.DB
}

func NewMasterService(db repo.DB) *MasterService { return &MasterService{db} }

func (s *MasterService) Categories(ctx context.Context, activeOnly bool) ([]model.Category, *apperr.Error) {
	out, err := repo.ListCategories(ctx, s.db, activeOnly)
	if err != nil {
		return nil, apperr.Internal()
	}
	return out, nil
}

func (s *MasterService) CreateCategory(ctx context.Context, nama, deskripsi string) (*model.Category, *apperr.Error) {
	if strings.TrimSpace(nama) == "" {
		return nil, apperr.BadRequest("NAMA_WAJIB", "Nama kategori wajib diisi.")
	}
	c, err := repo.CreateCategory(ctx, s.db, strings.TrimSpace(nama), optStr(deskripsi))
	if err != nil {
		return nil, apperr.FromPG(err, "kategori")
	}
	return c, nil
}

func (s *MasterService) UpdateCategory(ctx context.Context, id, nama, deskripsi string, active bool) (*model.Category, *apperr.Error) {
	if !model.IsUUID(id) || strings.TrimSpace(nama) == "" {
		return nil, apperr.BadRequest("INPUT_TIDAK_VALID", "Input kategori tidak valid.")
	}
	c, err := repo.UpdateCategory(ctx, s.db, id, strings.TrimSpace(nama), optStr(deskripsi), active)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.NotFound("KATEGORI_TIDAK_DITEMUKAN", "Kategori tidak ditemukan.")
		}
		return nil, apperr.FromPG(err, "kategori")
	}
	return c, nil
}

func (s *MasterService) DeleteCategory(ctx context.Context, id string) *apperr.Error {
	if !model.IsUUID(id) {
		return apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid.")
	}
	n, err := repo.CountAssetsByCategory(ctx, s.db, id)
	if err != nil {
		return apperr.Internal()
	}
	if n > 0 {
		return apperr.Conflict("KATEGORI_TERPAKAI", "Kategori masih dipakai aset. Nonaktifkan saja.")
	}
	if err := repo.DeleteCategory(ctx, s.db, id); err != nil {
		return apperr.FromPG(err, "kategori")
	}
	return nil
}

func (s *MasterService) Locations(ctx context.Context, activeOnly bool) ([]model.Location, *apperr.Error) {
	out, err := repo.ListLocations(ctx, s.db, activeOnly)
	if err != nil {
		return nil, apperr.Internal()
	}
	return out, nil
}

func (s *MasterService) CreateLocation(ctx context.Context, nama, gedung string) (*model.Location, *apperr.Error) {
	if strings.TrimSpace(nama) == "" {
		return nil, apperr.BadRequest("NAMA_WAJIB", "Nama lokasi wajib diisi.")
	}
	l, err := repo.CreateLocation(ctx, s.db, strings.TrimSpace(nama), optStr(gedung))
	if err != nil {
		return nil, apperr.FromPG(err, "lokasi")
	}
	return l, nil
}

func (s *MasterService) UpdateLocation(ctx context.Context, id, nama, gedung string, active bool) (*model.Location, *apperr.Error) {
	if !model.IsUUID(id) || strings.TrimSpace(nama) == "" {
		return nil, apperr.BadRequest("INPUT_TIDAK_VALID", "Input lokasi tidak valid.")
	}
	l, err := repo.UpdateLocation(ctx, s.db, id, strings.TrimSpace(nama), optStr(gedung), active)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.NotFound("LOKASI_TIDAK_DITEMUKAN", "Lokasi tidak ditemukan.")
		}
		return nil, apperr.FromPG(err, "lokasi")
	}
	return l, nil
}

func (s *MasterService) DeleteLocation(ctx context.Context, id string) *apperr.Error {
	if !model.IsUUID(id) {
		return apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid.")
	}
	n, err := repo.CountAssetsByLocation(ctx, s.db, id)
	if err != nil {
		return apperr.Internal()
	}
	if n > 0 {
		return apperr.Conflict("LOKASI_TERPAKAI", "Lokasi masih dipakai aset. Nonaktifkan saja.")
	}
	if err := repo.DeleteLocation(ctx, s.db, id); err != nil {
		return apperr.FromPG(err, "lokasi")
	}
	return nil
}

// ---- Users (admin) ----

type UserService struct {
	db repo.DB
}

func NewUserService(db repo.DB) *UserService { return &UserService{db} }

func (s *UserService) List(ctx context.Context, q, role string, active *bool, page, per int) ([]model.User, model.PageMeta, *apperr.Error) {
	if role != "" && !model.ValidRole(role) {
		return nil, model.PageMeta{}, apperr.BadRequest("ROLE_TIDAK_VALID", "Role tidak dikenal.")
	}
	items, total, err := repo.ListUsers(ctx, s.db, q, role, active, page, per)
	if err != nil {
		return nil, model.PageMeta{}, apperr.Internal()
	}
	return items, model.NewPageMeta(page, per, total), nil
}

type CreateUserInput struct {
	Nama     string `json:"nama"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (s *UserService) Create(ctx context.Context, in CreateUserInput) (*model.User, *apperr.Error) {
	in.Nama = strings.TrimSpace(in.Nama)
	in.Email = strings.TrimSpace(in.Email)
	if in.Nama == "" || in.Email == "" || len(in.Password) < 8 {
		return nil, apperr.BadRequest("INPUT_TIDAK_VALID", "Nama, email wajib diisi; password minimal 8 karakter.")
	}
	if !model.ValidRole(in.Role) {
		return nil, apperr.BadRequest("ROLE_TIDAK_VALID", "Role tidak dikenal.")
	}
	hash, herr := auth.HashPassword(in.Password)
	if herr != nil {
		return nil, apperr.Internal()
	}
	u, err := repo.CreateUser(ctx, s.db, in.Nama, in.Email, hash, in.Role)
	if err != nil {
		return nil, apperr.FromPG(err, "pengguna")
	}
	return u, nil
}

func (s *UserService) Update(ctx context.Context, id, nama, role string, active bool, actor model.ContextUser) (*model.User, *apperr.Error) {
	if !model.IsUUID(id) {
		return nil, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid.")
	}
	if !model.ValidRole(role) || strings.TrimSpace(nama) == "" {
		return nil, apperr.BadRequest("INPUT_TIDAK_VALID", "Nama dan role wajib valid.")
	}
	// Cegah admin menonaktifkan diri sendiri / admin terakhir.
	if id == actor.ID && !active {
		return nil, apperr.Conflict("NONAKTIF_SENDIRI", "Tidak bisa menonaktifkan akun sendiri.")
	}
	cur, err := repo.FindUserByID(ctx, s.db, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.NotFound("USER_TIDAK_DITEMUKAN", "Pengguna tidak ditemukan.")
		}
		return nil, apperr.Internal()
	}
	if cur.Role == model.RoleAdmin && (role != model.RoleAdmin || !active) {
		n, cerr := repo.CountAdmins(ctx, s.db)
		if cerr != nil {
			return nil, apperr.Internal()
		}
		if n <= 1 {
			return nil, apperr.Conflict("ADMIN_TERAKHIR", "Tidak bisa mencabut admin terakhir.")
		}
	}
	if !active {
		n, cerr := repo.CountAssetsByHolder(ctx, s.db, id)
		if cerr != nil {
			return nil, apperr.Internal()
		}
		if n > 0 {
			return nil, apperr.Conflict("HOLDER_AKTIF", "User masih memegang aset. Mutasikan dulu.")
		}
		_ = repo.RevokeAllUserSessions(ctx, s.db, id) // nonaktif langsung memutus sesi
	}
	u, uerr := repo.UpdateUser(ctx, s.db, id, strings.TrimSpace(nama), role, active)
	if uerr != nil {
		return nil, apperr.FromPG(uerr, "pengguna")
	}
	return u, nil
}

// ---- Dashboard ----

type DashboardService struct {
	pool TxBeginner
	db   repo.DB
}

func NewDashboardService(pool TxBeginner, db repo.DB) *DashboardService {
	return &DashboardService{pool, db}
}

func (s *DashboardService) Get(ctx context.Context, u model.ContextUser) (*model.Dashboard, *apperr.Error) {
	counts, err := repo.AssetStatusCounts(ctx, s.db)
	if err != nil {
		return nil, apperr.Internal()
	}
	open, err := repo.CountOpenTickets(ctx, s.db)
	if err != nil {
		return nil, apperr.Internal()
	}
	d := &model.Dashboard{
		Available:     counts[model.AssetAvailable],
		Assigned:      counts[model.AssetAssigned],
		InMaintenance: counts[model.AssetMaintenance],
		Inactive:      counts[model.AssetInactive],
		OpenTickets:   open,
		RecentMoves:   []model.AssetMovement{},
		ActionTickets: []model.Maintenance{},
	}
	d.TotalAssets = d.Available + d.Assigned + d.InMaintenance + d.Inactive
	if u.Role == model.RoleKaryawan {
		d.MyAssets, _ = repo.CountMyAssets(ctx, s.db, u.ID)
		d.MyOpenTickets, _ = repo.CountMyOpenTickets(ctx, s.db, u.ID)
		moves, merr := repo.RecentMovements(ctx, s.db, 8, u.ID)
		if merr == nil {
			d.RecentMoves = moves
		}
		tix, terr := repo.OpenTicketsForAction(ctx, s.db, u.ID, true, 8)
		if terr == nil {
			d.ActionTickets = tix
		}
	} else {
		moves, merr := repo.RecentMovements(ctx, s.db, 8, "")
		if merr == nil {
			d.RecentMoves = moves
		}
		tix, terr := repo.OpenTicketsForAction(ctx, s.db, "", false, 8)
		if terr == nil {
			d.ActionTickets = tix
		}
	}
	return d, nil
}
