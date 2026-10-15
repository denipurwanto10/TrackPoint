// Package server menyediakan HTTP server Gin dengan routing dan middleware dasar.
package server

import (
	"context"
	"net/http"
	"time"

	"trackpoint/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	router *gin.Engine
	server *http.Server
	cfg    config.Config
	pg     *pgxpool.Pool // koneksi pool database (bisa nil jika DB tidak tersedia)
}

func New(cfg config.Config, pg *pgxpool.Pool) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	s := &Server{
		router: r,
		cfg:    cfg,
		pg:     pg,
	}
	s.setupRoutes()
	s.setupServer()
	return s
}

func (s *Server) setupRoutes() {
	// CORS middleware - only allow frontend origin
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", s.cfg.CORSOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check - does NOT check database, only that server is running
	s.router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API v1 routes group (will be expanded in next phases)
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})

		// Readiness: pemeriksaan kesiapan DB tanpa membocorkan detail koneksi.
		v1.GET("/ready", func(c *gin.Context) {
			if s.pg == nil {
				// Database belum terkonfigurasi — server masih sehat untuk dev lokal.
				c.JSON(http.StatusOK, gin.H{
					"status":    "ready",
					"database":  "unconfigured",
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
			defer cancel()

			err := s.pg.Ping(ctx)
			if err != nil {
				// Tidak kirim error message ke klien — hanya kode status.
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status":    "degraded",
					"database":  "unavailable",
					"timestamp": time.Now().UTC().Format(time.RFC3339),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"status":    "ready",
				"database":  "connected",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
		})
	}
}

func (s *Server) setupServer() {
	s.server = &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
