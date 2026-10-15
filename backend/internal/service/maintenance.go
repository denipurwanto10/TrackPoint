// Service maintenance: tiket, transisi status, sinkron status aset.
package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"trackpoint/internal/apperr"
	"trackpoint/internal/model"
	"trackpoint/internal/repo"
)

type MaintenanceService struct {
	pool TxBeginner
	db   repo.DB
}

func NewMaintenanceService(pool TxBeginner, db repo.DB) *MaintenanceService {
	return &MaintenanceService{pool, db}
}

func (s *MaintenanceService) List(ctx context.Context, status string, u model.ContextUser, page, per int) ([]model.Maintenance, model.PageMeta, *apperr.Error) {
	mine := ""
	if u.Role == model.RoleKaryawan {
		mine = u.ID // karyawan hanya lihat tiket miliknya
	}
	items, total, err := repo.ListMaintenance(ctx, s.db, status, mine, page, per)
	if err != nil {
		return nil, model.PageMeta{}, apperr.Internal()
	}
	return items, model.NewPageMeta(page, per, total), nil
}

func (s *MaintenanceService) Get(ctx context.Context, id string, u model.ContextUser) (*model.Maintenance, *apperr.Error) {
	if !model.IsUUID(id) {
		return nil, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid.")
	}
	m, err := repo.GetMaintenance(ctx, s.db, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperr.NotFound("TIKET_TIDAK_DITEMUKAN", "Tiket tidak ditemukan.")
		}
		return nil, apperr.Internal()
	}
	if u.Role == model.RoleKaryawan && m.ReportedBy != u.ID {
		return nil, apperr.Forbidden("AKSES_DITOLAK", "Anda hanya bisa melihat tiket sendiri.")
	}
	return m, nil
}

// Report: karyawan hanya untuk aset yang sedang dipegangnya; teknisi/admin bebas.
func (s *MaintenanceService) Report(ctx context.Context, assetID, keluhan string, u model.ContextUser) (*model.Maintenance, *apperr.Error) {
	if !model.IsUUID(assetID) || strings.TrimSpace(keluhan) == "" {
		return nil, apperr.BadRequest("INPUT_TIDAK_VALID", "Aset dan keluhan wajib diisi.")
	}
	var out *model.Maintenance
	err := withTx(ctx, s.pool, func(tx pgx.Tx) error {
		a, gerr := repo.GetAssetForUpdate(ctx, tx, assetID)
		if gerr != nil {
			return gerr
		}
		if u.Role == model.RoleKaryawan {
			if a.HolderID == nil || *a.HolderID != u.ID {
				return errNotHolder
			}
		}
		m, merr := repo.CreateMaintenance(ctx, tx, assetID, a.Kode, strings.TrimSpace(keluhan), u.ID)
		if merr != nil {
			return merr
		}
		// Aset masuk maintenance (kecuali sudah di sana).
		if a.Status != model.AssetMaintenance {
			if merr := repo.ApplyAssetMove(ctx, tx, assetID, model.AssetMaintenance, a.LocationID, nil); merr != nil {
				return merr
			}
		}
		out = m
		return nil
	})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, apperr.NotFound("ASET_TIDAK_DITEMUKAN", "Aset tidak ditemukan.")
		case errNotHolder:
			return nil, apperr.Forbidden("BUKAN_PEMEGANG", "Hanya bisa melaporkan aset yang ditugaskan kepadamu.")
		}
		return nil, apperr.FromPG(err, "tiket")
	}
	return s.Get(ctx, out.ID, u)
}

type ProgressInput struct {
	TechnicianID string
	Tindakan     string
	Biaya        *float64
	ReturnTo     string
	CancelReason string
}

// Transition validasi peran + urutan status + kolom pendukung, semua dalam tx.
func (s *MaintenanceService) Transition(ctx context.Context, id, to string, in ProgressInput, u model.ContextUser) (*model.Maintenance, *apperr.Error) {
	if !model.IsUUID(id) {
		return nil, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid.")
	}
	if u.Role == model.RoleKaryawan {
		return nil, apperr.Forbidden("AKSES_DITOLAK", "Hanya teknisi atau admin yang memproses tiket.")
	}
	var out *model.Maintenance
	err := withTx(ctx, s.pool, func(tx pgx.Tx) error {
		m, gerr := repo.GetMaintenanceForUpdate(ctx, tx, id)
		if gerr != nil {
			return gerr
		}
		if !model.CanTransitionMaintenance(m.Status, to) {
			return errBadTransition
		}
		tech := in.TechnicianID
		if tech == "" {
			tech = coalesceStr(m.TechnicianID, u.ID)
		}
		if to == model.MntInProgress && tech == "" {
			return errNeedTech
		}
		if in.Biaya != nil && *in.Biaya < 0 {
			return apperr.BadRequest("BIAYA_NEGATIF", "Biaya tidak boleh negatif")
		}
		if to == model.MntDone {
			if strings.TrimSpace(in.Tindakan) == "" || in.ReturnTo == "" {
				return errNeedDoneFields
			}
			if in.ReturnTo != model.ReturnAvailable && in.ReturnTo != model.ReturnAssigned && in.ReturnTo != model.ReturnInactive {
				return errBadReturnTo
			}
		}
		if to == model.MntCancelled && strings.TrimSpace(in.CancelReason) == "" {
			return errNeedCancelReason
		}
		updated, uerr := repo.UpdateMaintenanceStatus(ctx, tx, id, to, strings.TrimSpace(in.Tindakan), in.Biaya, in.ReturnTo, strings.TrimSpace(in.CancelReason), tech)
		if uerr != nil {
			return uerr
		}
		out = updated
		if to == model.MntDone {
			// Sinkron status aset sesuai returnTo.
			holder, herr := assetHolderForReturn(in.ReturnTo, tech)
			if herr != nil {
				return herr
			}
			// Lock asset row to prevent race with concurrent Move
			if _, err := repo.GetAssetForUpdate(ctx, tx, m.AssetID); err != nil {
				return err
			}
			loc, lerr := assetLoc(ctx, tx, m.AssetID)
			if lerr != nil {
				return lerr
			}
			if serr := repo.ApplyAssetMove(ctx, tx, m.AssetID, in.ReturnTo, loc, holder); serr != nil {
				return serr
			}
		}
		if to == model.MntCancelled {
			// Batal → kembalikan aset ke AVAILABLE
			// Lock asset row to prevent race with concurrent Move
			if _, err := repo.GetAssetForUpdate(ctx, tx, m.AssetID); err != nil {
				return err
			}
			loc, lerr := assetLoc(ctx, tx, m.AssetID)
			if lerr != nil {
				return lerr
			}
			if serr := repo.ApplyAssetMove(ctx, tx, m.AssetID, model.ReturnAvailable, loc, nil); serr != nil {
				return serr
			}
		}
		return nil
	})
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return nil, apperr.NotFound("TIKET_TIDAK_DITEMUKAN", "Tiket tidak ditemukan.")
		case errBadTransition:
			return nil, apperr.Unprocessable("TRANSISI_DITOLAK", "Perpindahan status tidak diizinkan.")
		case errNeedTech:
			return nil, apperr.BadRequest("TEKNISI_WAJIB", "Teknisi penanggung jawab wajib diisi.")
		case errNeedDoneFields:
			return nil, apperr.BadRequest("DATA_SELESAI_WAJIB", "Tindakan dan status akhir aset (returnTo) wajib diisi.")
		case errBadReturnTo:
			return nil, apperr.BadRequest("RETURN_TO_SALAH", "returnTo harus AVAILABLE, ASSIGNED, atau INACTIVE.")
		case errNeedCancelReason:
			return nil, apperr.BadRequest("ALASAN_BATAL_WAJIB", "Alasan pembatalan wajib diisi.")
		}
		return nil, apperr.FromPG(err, "tiket")
	}
	return out, nil
}

var (
	errNotHolder        = errSentinel("not-holder")
	errBadTransition    = errSentinel("bad-transition")
	errNeedTech         = errSentinel("need-tech")
	errNeedDoneFields   = errSentinel("need-done")
	errBadReturnTo      = errSentinel("bad-returnto")
	errNeedCancelReason = errSentinel("need-cancel")
)

func coalesceStr(p *string, fb string) string {
	if p != nil && *p != "" {
		return *p
	}
	return fb
}

// assetLoc baca lokasi terkini aset dalam tx yang sama.
// Returns error if query fails or asset not found.
func assetLoc(ctx context.Context, db repo.DB, assetID string) (string, error) {
	rows, err := db.Query(ctx, `SELECT location_id FROM assets WHERE id=$1`, assetID)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		var loc string
		if err := rows.Scan(&loc); err == nil {
			return loc, nil
		}
	}
	return "", pgx.ErrNoRows
}

// assetHolderForReturn: ASSIGNED → kembali ke teknisi penangan; lainnya → tanpa holder.
// Returns error if ASSIGNED but no tech provided (DB CHECK constraint requires holder_id).
func assetHolderForReturn(returnTo, tech string) (*string, error) {
	if returnTo == model.ReturnAssigned {
		if tech == "" {
			return nil, errNeedTech
		}
		return &tech, nil
	}
	return nil, nil
}
