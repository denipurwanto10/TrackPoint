// Routing API v1: wiring Deps + middleware auth/RBAC.
// Kesehatan (/api/health, /api/v1/health, /api/v1/ready) tetap di server.go.
package server

import (
	"trackpoint/internal/api"
	"trackpoint/internal/model"
	"trackpoint/internal/mw"
)

// RegisterAPI memasang seluruh endpoint aplikasi ke grup /api/v1.
func (s *Server) RegisterAPI(d *api.Deps) {
	v1 := s.router.Group("/api/v1")

	// Publik + cookie (refresh/logout dijaga origin check).
	auth := v1.Group("/auth")
	{
		auth.POST("/login", d.Login)
		auth.POST("/refresh", mw.CheckOrigin(s.cfg), d.Refresh)
		auth.POST("/logout", mw.CheckOrigin(s.cfg), d.Logout)
	}

	// Terproteksi.
	authed := v1.Group("", mw.Auth(s.cfg, s.pg))
	{
		authed.GET("/auth/me", d.Me)
		authed.GET("/auth/sessions", d.MySessions)
		authed.DELETE("/auth/sessions/:id", d.RevokeSession)

		authed.GET("/dashboard", d.GetDashboard)

		// Aset: baca semua role; tulis admin+teknisi.
		authed.GET("/assets", d.ListAssets)
		authed.GET("/assets/:id", d.GetAsset)
		authed.GET("/assets/:id/movements", d.ListMovements)
		write := authed.Group("", mw.RequireRole(model.RoleAdmin, model.RoleTeknisi))
		{
			write.POST("/assets", d.CreateAsset)
			write.PATCH("/assets/:id", d.UpdateAsset)
			write.POST("/assets/:id/movements", d.MoveAsset)
		}

		// Master: baca semua role; tulis admin saja.
		authed.GET("/categories", d.ListCategories)
		authed.GET("/locations", d.ListLocations)

		// Maintenance: lapor semua role (service batasi karyawan ke aset sendiri);
		// transisi admin+teknisi.
		authed.GET("/maintenance", d.ListMaintenance)
		authed.GET("/maintenance/:id", d.GetMaintenance)
		authed.POST("/maintenance", d.ReportMaintenance)
		authed.PATCH("/maintenance/:id", mw.RequireRole(model.RoleAdmin, model.RoleTeknisi), d.TransitionMaintenance)

		// Admin penuh.
		admin := authed.Group("", mw.RequireRole(model.RoleAdmin))
		{
			admin.POST("/categories", d.CreateCategory)
			admin.PATCH("/categories/:id", d.UpdateCategory)
			admin.DELETE("/categories/:id", d.DeleteCategory)
			admin.POST("/locations", d.CreateLocation)
			admin.PATCH("/locations/:id", d.UpdateLocation)
			admin.DELETE("/locations/:id", d.DeleteLocation)

			admin.GET("/users", d.ListUsers)
			admin.POST("/users", d.CreateUser)
			admin.PATCH("/users/:id", d.UpdateUser)
			admin.DELETE("/sessions/:id", d.AdminRevokeSession)
		}
	}
}
