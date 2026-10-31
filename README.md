# TrackPoint — IT Asset & Maintenance Management (lokal) 
 
Aplikasi web lokal untuk kelola aset IT: aset, lokasi, penanggung jawab, histori mutasi, dan maintenance. Sepenuhnya berjalan di `localhost` Windows. Tanpa cloud, tanpa API key, tanpa biaya berlangganan.

## Status

Backend lengkap: auth (login/refresh rotation + reuse detection/logout, cookie HttpOnly `tp_refresh`), RBAC ADMIN/TEKNISI/KARYAWAN di tiap endpoint, dashboard agregat dari DB, CRUD aset (kode `TRK-000001` via `asset_code_seq`, mutasi transaksional append-only), maintenance (REPORTED→IN_PROGRESS→DONE/CANCELLED), users + master kategori/lokasi. Frontend lengkap: login, dashboard, aset + mutasi, maintenance, sesi saya, master (ADMIN), pengguna (ADMIN). Guard `trackpoint_test` aktif; kontrak JSON snake_case, login+refresh sama-sama `access_token`.

Terverifikasi 2026-10-15: `gofmt` bersih, `go vet` bersih, `go build ./...` OK, `go test` lolos, `tsc --noEmit` bersih, `vite build` sukses (tanpa warning). **Belum terverifikasi**: migration Goose + seed + alur E2E — PostgreSQL/`psql` tidak tersedia di mesin ini. Jangan klaim siap produksi sebelum itu diuji.

## Prasyarat (install sekali, gratis)

- Go 1.24+ (terpasang di mesin ini: go1.27.0)
- Node.js LTS (untuk frontend & Playwright, tahap berikutnya)
- PostgreSQL lokal dengan `psql` di PATH
- PowerShell

Docker opsional, bukan syarat.

## Konvensi lokal

| Bagian | Alamat |
|---|---|
| Frontend | http://localhost:5173 |
| Backend | http://localhost:8080 |
| PostgreSQL | localhost:5432 |
| DB pengembangan | `trackpoint` |
| DB pengujian | `trackpoint_test` |

## Setup

```powershell
pwsh -File scripts/setup.ps1      # buat .env dari .env.example + generate secret JWT lokal (aman diulang)
# edit .env: isi password DB + superuser
pwsh -File scripts/dbsetup.ps1    # buat database trackpoint & trackpoint_test bila belum ada (aman diulang)
cd backend
go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres "$env:DATABASE_URL" up
# opsional: seed admin awal (password acak tampil SEKALI di stdout)
go run ./cmd/seed
```

Cara jalan lokal (dua terminal):

```powershell
cd backend; go run ./cmd/api          # :8080
cd frontend; npm run dev              # :5173 (proxy /api → :8080)
```

`dbsetup.ps1` tidak mengubah `pg_hba.conf` dan tidak memakai autentikasi `trust` — password SCRAM tetap dipakai. Kalau `psql` tidak ada di PATH, jalankan lewat "SQL Shell (psql)" dari menu Start PostgreSQL.

## Verifikasi aman (tidak menyentuh database)

```powershell
cd backend
go build ./...
go vet ./...
go test ./...
```

## Pengujian database

Integration test memakai guard di `internal/testdb`: test **abort** bila `DATABASE_URL` menunjuk database selain `trackpoint_test`, jadi test tidak pernah memutasi database pengembangan. Migration juga tidak menyediakan bagian `Down` untuk mencegah reset destruktif.

## Struktur

```
backend/
  cmd/api/        entry point server (main.go)
  cmd/seed/       seed admin awal + master minimal (password acak, tampil sekali)
  internal/
    config/       muat & validasi konfigurasi dari environment
    db/           pool PostgreSQL via pgxpool (NewPool, ClosePool)
    server/       router Gin + wiring API, health check, readiness DB
    api/          handler HTTP + cookie refresh (tp_refresh)
    mw/           Auth JWT, RequireRole, CheckOrigin
    auth/         JWT access 15 mnt + bcrypt + refresh opaque
    model/        tipe domain + validasi
    repo/         query SQL (transaksi, rotate conditional, append-only)
    service/      aturan bisnis per fitur
    apperr/       error terstruktur {error:{code,message}}
    testdb/       guard database pengujian
  migrations/     goose, hanya bagian Up
frontend/
  src/api.ts      token di memori modul + single-flight refresh + retry-401-sekali
  src/auth.tsx    konteks sesi (login/logout/restore via cookie)
  src/pages/      Dashboard, Assets, Maintenance, Master, Users, Sessions
scripts/          setup.ps1, dbsetup.ps1
.env.example      template tanpa secret
```

## Catatan keamanan

- Access JWT hanya di memori frontend (Bearer). Refresh token hanya di cookie HttpOnly (`tp_refresh`, SameSite=Lax, Path=/api/v1/auth).
- `Secure=false` pada cookie hanya untuk HTTP localhost. **Tidak aman untuk produksi** — produksi butuh HTTPS + `Secure=true`.
- `.env` tidak pernah di-commit; `.env.example` tidak berisi secret asli.
- Refresh token disimpan sebagai hash (SHA-256) di PostgreSQL dengan rotation dan reuse detection.
