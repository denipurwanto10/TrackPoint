// Package api: handler HTTP Gin. Semua error memakai {error:{code,message}}.
package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"trackpoint/internal/apperr"
	"trackpoint/internal/config"
	"trackpoint/internal/model"
	"trackpoint/internal/mw"
	"trackpoint/internal/repo"
	"trackpoint/internal/service"
)

type Deps struct {
	Cfg         config.Config
	DB          repo.DB
	Auth        *service.AuthService
	Assets      *service.AssetService
	Maintenance *service.MaintenanceService
	Master      *service.MasterService
	Users       *service.UserService
	Dashboard   *service.DashboardService
}

func fail(c *gin.Context, e *apperr.Error) {
	c.JSON(e.Status, gin.H{"error": gin.H{"code": e.Code, "message": e.Message}})
}

func ok(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"data": data})
}

func okPage(c *gin.Context, data any, meta model.PageMeta) {
	c.JSON(http.StatusOK, gin.H{"data": data, "meta": meta})
}

func paging(c *gin.Context) (int, int) {
	return model.ParsePagination(c.Query("page"), c.Query("per_page"))
}

func clientIP(c *gin.Context) string { return c.ClientIP() }

// ---- Auth ----

func (d *Deps) Login(c *gin.Context) {
	var in struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Email dan password wajib diisi."))
		return
	}
	res, aerr := d.Auth.Login(c.Request.Context(), in.Email, in.Password, c.GetHeader("User-Agent"), clientIP(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	setRefreshCookie(c, d.Cfg, res.Refresh, int(res.ExpiresAt.Unix()-nowUnix()))
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"access_token": res.AccessToken,
		"user":         res.User,
	}})
}

func (d *Deps) Refresh(c *gin.Context) {
	raw, _ := c.Cookie("tp_refresh")
	res, aerr := d.Auth.Refresh(c.Request.Context(), raw, c.GetHeader("User-Agent"), clientIP(c))
	if aerr != nil {
		clearRefreshCookie(c, d.Cfg)
		fail(c, aerr)
		return
	}
	setRefreshCookie(c, d.Cfg, res.Refresh, int(res.ExpiresAt.Unix()-nowUnix()))
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"access_token": res.AccessToken, "user": res.User}})
}

func (d *Deps) Logout(c *gin.Context) {
	raw, _ := c.Cookie("tp_refresh")
	if aerr := d.Auth.Logout(c.Request.Context(), raw); aerr != nil {
		fail(c, aerr)
		return
	}
	clearRefreshCookie(c, d.Cfg)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"ok": true}})
}

func (d *Deps) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": mustUser(c)})
}

func (d *Deps) MySessions(c *gin.Context) {
	u := mustUser(c)
	items, err := repo.ListActiveSessions(c.Request.Context(), d.DB, u.ID)
	if err != nil {
		fail(c, apperr.Internal())
		return
	}
	ok(c, http.StatusOK, items)
}

func (d *Deps) RevokeSession(c *gin.Context) {
	u := mustUser(c)
	id := c.Param("id")
	if !model.IsUUID(id) {
		fail(c, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid."))
		return
	}
	// User hanya boleh mencabut sesi miliknya; admin boleh semua.
	items, err := repo.ListActiveSessions(c.Request.Context(), d.DB, u.ID)
	if err != nil {
		fail(c, apperr.Internal())
		return
	}
	own := false
	for _, s := range items {
		if s.ID == id {
			own = true
		}
	}
	if !own && u.Role != model.RoleAdmin {
		fail(c, apperr.Forbidden("AKSES_DITOLAK", "Anda tidak punya akses mencabut sesi ini."))
		return
	}
	if err := repo.RevokeSession(c.Request.Context(), d.DB, id); err != nil {
		fail(c, apperr.Internal())
		return
	}
	ok(c, http.StatusOK, gin.H{"ok": true})
}

func mustUser(c *gin.Context) model.ContextUser {
	u, _ := mw.Current(c)
	return u
}

// ---- Dashboard ----

func (d *Deps) GetDashboard(c *gin.Context) {
	u := mustUser(c)
	res, aerr := d.Dashboard.Get(c.Request.Context(), u)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, res)
}

// ---- Assets ----

func (d *Deps) ListAssets(c *gin.Context) {
	page, per := paging(c)
	u := mustUser(c)
	items, meta, aerr := d.Assets.List(c.Request.Context(), c.Query("q"), c.Query("status"),
		c.Query("category_id"), c.Query("location_id"), c.Query("holder_id"), page, per, &u)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	okPage(c, items, meta)
}

func (d *Deps) GetAsset(c *gin.Context) {
	a, aerr := d.Assets.Get(c.Request.Context(), c.Param("id"))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, a)
}

func (d *Deps) CreateAsset(c *gin.Context) {
	var in service.CreateAssetInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data aset tidak valid."))
		return
	}
	a, aerr := d.Assets.Create(c.Request.Context(), in, mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": a})
}

func (d *Deps) UpdateAsset(c *gin.Context) {
	var in service.UpdateAssetInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data aset tidak valid."))
		return
	}
	a, aerr := d.Assets.UpdateMeta(c.Request.Context(), c.Param("id"), in, mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, a)
}

func (d *Deps) MoveAsset(c *gin.Context) {
	var in service.MoveInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data mutasi tidak valid."))
		return
	}
	m, aerr := d.Assets.Move(c.Request.Context(), c.Param("id"), in, mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": m})
}

func (d *Deps) ListMovements(c *gin.Context) {
	if !model.IsUUID(c.Param("id")) {
		fail(c, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid."))
		return
	}
	page, per := paging(c)
	items, total, err := repo.ListMovements(c.Request.Context(), d.DB, c.Param("id"), page, per)
	if err != nil {
		fail(c, apperr.Internal())
		return
	}
	okPage(c, items, model.NewPageMeta(page, per, total))
}

// ---- Master ----

func (d *Deps) ListCategories(c *gin.Context) {
	items, aerr := d.Master.Categories(c.Request.Context(), c.Query("active") == "true")
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, items)
}

func (d *Deps) CreateCategory(c *gin.Context) {
	var in struct {
		Nama      string `json:"nama"`
		Deskripsi string `json:"deskripsi"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Nama kategori wajib diisi."))
		return
	}
	out, aerr := d.Master.CreateCategory(c.Request.Context(), in.Nama, in.Deskripsi)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": out})
}

func (d *Deps) UpdateCategory(c *gin.Context) {
	var in struct {
		Nama      string `json:"nama"`
		Deskripsi string `json:"deskripsi"`
		IsActive  *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data kategori tidak valid."))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	out, aerr := d.Master.UpdateCategory(c.Request.Context(), c.Param("id"), in.Nama, in.Deskripsi, active)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, out)
}

func (d *Deps) DeleteCategory(c *gin.Context) {
	if aerr := d.Master.DeleteCategory(c.Request.Context(), c.Param("id")); aerr != nil {
		fail(c, aerr)
		return
	}
	c.Status(http.StatusNoContent)
}

func (d *Deps) ListLocations(c *gin.Context) {
	items, aerr := d.Master.Locations(c.Request.Context(), c.Query("active") == "true")
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, items)
}

func (d *Deps) CreateLocation(c *gin.Context) {
	var in struct {
		Nama   string `json:"nama"`
		Gedung string `json:"gedung"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Nama lokasi wajib diisi."))
		return
	}
	out, aerr := d.Master.CreateLocation(c.Request.Context(), in.Nama, in.Gedung)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": out})
}

func (d *Deps) UpdateLocation(c *gin.Context) {
	var in struct {
		Nama     string `json:"nama"`
		Gedung   string `json:"gedung"`
		IsActive *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data lokasi tidak valid."))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	out, aerr := d.Master.UpdateLocation(c.Request.Context(), c.Param("id"), in.Nama, in.Gedung, active)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, out)
}

func (d *Deps) DeleteLocation(c *gin.Context) {
	if aerr := d.Master.DeleteLocation(c.Request.Context(), c.Param("id")); aerr != nil {
		fail(c, aerr)
		return
	}
	c.Status(http.StatusNoContent)
}

// ---- Maintenance ----

func (d *Deps) ListMaintenance(c *gin.Context) {
	page, per := paging(c)
	items, meta, aerr := d.Maintenance.List(c.Request.Context(), c.Query("status"), mustUser(c), page, per)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	okPage(c, items, meta)
}

func (d *Deps) GetMaintenance(c *gin.Context) {
	m, aerr := d.Maintenance.Get(c.Request.Context(), c.Param("id"), mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, m)
}

func (d *Deps) ReportMaintenance(c *gin.Context) {
	var in struct {
		AssetID string `json:"asset_id"`
		Keluhan string `json:"keluhan"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Aset dan keluhan wajib diisi."))
		return
	}
	m, aerr := d.Maintenance.Report(c.Request.Context(), in.AssetID, in.Keluhan, mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": m})
}

func (d *Deps) TransitionMaintenance(c *gin.Context) {
	var in struct {
		To           string   `json:"to"`
		TechnicianID string   `json:"technician_id"`
		Tindakan     string   `json:"tindakan"`
		Biaya        *float64 `json:"biaya"`
		ReturnTo     string   `json:"return_to"`
		CancelReason string   `json:"cancel_reason"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || in.To == "" {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Status tujuan wajib diisi."))
		return
	}
	m, aerr := d.Maintenance.Transition(c.Request.Context(), c.Param("id"), in.To, service.ProgressInput{
		TechnicianID: in.TechnicianID, Tindakan: in.Tindakan, Biaya: in.Biaya,
		ReturnTo: in.ReturnTo, CancelReason: in.CancelReason,
	}, mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, m)
}

// ---- Users ----

func (d *Deps) ListUsers(c *gin.Context) {
	var active *bool
	if v := c.Query("is_active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			active = &b
		}
	}
	page, per := paging(c)
	items, meta, aerr := d.Users.List(c.Request.Context(), c.Query("q"), c.Query("role"), active, page, per)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	okPage(c, items, meta)
}

func (d *Deps) CreateUser(c *gin.Context) {
	var in service.CreateUserInput
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data pengguna tidak valid."))
		return
	}
	// Email lowercase-kan agar konsisten dengan unique index lower(email).
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	u, aerr := d.Users.Create(c.Request.Context(), in)
	if aerr != nil {
		fail(c, aerr)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": u})
}

func (d *Deps) UpdateUser(c *gin.Context) {
	var in struct {
		Nama     string `json:"nama"`
		Role     string `json:"role"`
		IsActive *bool  `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		fail(c, apperr.BadRequest("INPUT_TIDAK_VALID", "Data pengguna tidak valid."))
		return
	}
	active := true
	if in.IsActive != nil {
		active = *in.IsActive
	}
	u, aerr := d.Users.Update(c.Request.Context(), c.Param("id"), in.Nama, in.Role, active, mustUser(c))
	if aerr != nil {
		fail(c, aerr)
		return
	}
	ok(c, http.StatusOK, u)
}

func (d *Deps) AdminRevokeSession(c *gin.Context) {
	if !model.IsUUID(c.Param("id")) {
		fail(c, apperr.BadRequest("FORMAT_ID_SALAH", "Format ID tidak valid."))
		return
	}
	if err := repo.RevokeSession(c.Request.Context(), d.DB, c.Param("id")); err != nil {
		fail(c, apperr.Internal())
		return
	}
	ok(c, http.StatusOK, gin.H{"ok": true})
}
