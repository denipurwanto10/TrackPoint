package model

import "testing"

func TestValidAssetStatus(t *testing.T) {
	for _, s := range []string{AssetAvailable, AssetAssigned, AssetMaintenance, AssetInactive} {
		if !ValidAssetStatus(s) {
			t.Errorf("%s harus valid", s)
		}
	}
	if ValidAssetStatus("RUSAK") || ValidAssetStatus("available") {
		t.Error("status asing/casing salah diterima")
	}
}

func TestMaintenanceTransitions(t *testing.T) {
	if !CanTransitionMaintenance(MntReported, MntInProgress) {
		t.Error("reported→in_progress harus boleh")
	}
	if !CanTransitionMaintenance(MntReported, MntCancelled) {
		t.Error("reported→cancelled harus boleh")
	}
	if CanTransitionMaintenance(MntReported, MntDone) {
		t.Error("reported→done langsung harus ditolak")
	}
	if CanTransitionMaintenance(MntDone, MntInProgress) {
		t.Error("done→in_progress harus ditolak")
	}
	if !CanTransitionMaintenance(MntInProgress, MntDone) {
		t.Error("in_progress→done harus boleh")
	}
	// returnTo hanya tiga nilai ini (dipakai service maintenance).
	for _, r := range []string{ReturnAvailable, ReturnAssigned, ReturnInactive} {
		if !ValidAssetStatus(r) {
			t.Errorf("returnTo %s harus status aset valid", r)
		}
	}
}

func TestRoles(t *testing.T) {
	if !ValidRole(RoleAdmin) || !ValidRole(RoleTeknisi) || !ValidRole(RoleKaryawan) {
		t.Error("role baku ditolak")
	}
	if ValidRole("SUPERADMIN") || ValidRole("admin") {
		t.Error("role asing/casing salah diterima")
	}
	if !IsUUID("123e4567-e89b-12d3-a456-426614174000") || IsUUID("bukan-uuid") {
		t.Error("validasi UUID salah")
	}
}

func TestPagination(t *testing.T) {
	p, per := ParsePagination("0", "9999")
	if p != 1 || per != 20 {
		t.Errorf("clamp gagal: %d/%d", p, per)
	}
	p, per = ParsePagination("3", "10")
	if p != 3 || per != 10 {
		t.Errorf("normal gagal: %d/%d", p, per)
	}
	m := NewPageMeta(2, 10, 25)
	if m.TotalPages != 3 || m.Total != 25 {
		t.Errorf("meta salah: %+v", m)
	}
}
