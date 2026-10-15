// cmd/api adalah entry point server TrackPoint.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"trackpoint/internal/api"
	"trackpoint/internal/config"
	"trackpoint/internal/db"
	"trackpoint/internal/server"
	"trackpoint/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// DATABASE_URL wajib (divalidasi config.Load); tanpa DB server tidak start.
	var pg *pgxpool.Pool

	poolCfg := db.DefaultConfig()
	poolCfg.URL = cfg.DatabaseURL

	initPool := func() (*pgxpool.Pool, error) {
		poolCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return db.NewPool(poolCtx, poolCfg)
	}

	p, err := initPool()
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	pg = p
	log.Println("pool DB terinisialisasi")

	srv := server.New(cfg, pg)
	srv.RegisterAPI(&api.Deps{
		Cfg:         cfg,
		DB:          pg,
		Auth:        service.NewAuthService(pg, pg, cfg),
		Assets:      service.NewAssetService(pg, pg),
		Maintenance: service.NewMaintenanceService(pg, pg),
		Master:      service.NewMasterService(pg),
		Users:       service.NewUserService(pg),
		Dashboard:   service.NewDashboardService(pg, pg),
	})

	// Graceful shutdown: stop HTTP dulu, baru tutup pool.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}

	if err := db.ClosePool(pg); err != nil {
		log.Printf("warning: tutup pool DB: %v", err)
	}
	log.Println("server stopped")
}
