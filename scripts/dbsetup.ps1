#!/usr/bin/env pwsh
# Membuat role aplikasi + database trackpoint dan trackpoint_test bila belum ada (idempoten).
# TIDAK mengubah pg_hba.conf, TIDAK memakai auth trust, TIDAK menghapus database yang ada.
# TIDAK mengubah password role yang sudah ada: role existing dibiarkan apa adanya.
# Pakai:  pwsh -File scripts/dbsetup.ps1  (jalan juga di powershell.exe 5.1)
# Encoding: UTF-8 with BOM + CRLF agar terbaca benar di PowerShell 5.1.

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root '.env'

if (-not (Test-Path $envFile)) { throw '.env belum ada - jalankan scripts/setup.ps1 dulu.' }

# Parse sederhana key=value dari .env
$cfg = @{}
Get-Content $envFile | ForEach-Object {
    if ($_ -match '^\s*([A-Z_]+)\s*=\s*(.*)$') { $cfg[$matches[1]] = $matches[2].Trim() }
}

$super = $cfg['DB_SUPERUSER_URL']
$appUrl = $cfg['DATABASE_URL']
$dbName = $cfg['DB_NAME']
$testName = $cfg['DB_TEST_NAME']
if (-not $super -or $super -match 'GANTI_PASSWORD') { throw 'DB_SUPERUSER_URL belum diisi benar di .env (password masih placeholder).' }
if (-not $appUrl -or $appUrl -match 'GANTI_PASSWORD') { throw 'DATABASE_URL belum diisi benar di .env (password masih placeholder).' }
if (-not $dbName -or -not $testName) { throw 'DB_NAME / DB_TEST_NAME belum diisi di .env.' }
# Nama DB dan role diinterpolasi ke SQL, jadi batasi ke karakter aman.
foreach ($n in @($dbName, $testName)) {
    if ($n -notmatch '^[A-Za-z0-9_]+$') { throw "Nama database tidak valid: $n (hanya huruf/angka/underscore)." }
}

if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
    throw 'psql tidak ada di PATH. Buka "SQL Shell (psql)" dari menu Start PostgreSQL, atau tambahkan folder bin PostgreSQL ke PATH untuk sesi ini.'
}

# Urai kedua URL agar password tidak tampil di argumen proses (hanya lewat PGPASSWORD).
$sm = [regex]::Match($super, '^postgres(?:ql)?://([^:/?#]+)(?::([^@/?#]*))?@([^:/?#]+)(?::(\d+))?/([^?#]*)')
if (-not $sm.Success -or -not $sm.Groups[1].Value -or -not $sm.Groups[3].Value) {
    throw 'DB_SUPERUSER_URL tidak bisa diurai (format: postgres://user:pass@host:port/db).'
}
$am = [regex]::Match($appUrl, '^postgres(?:ql)?://([^:/?#]+)(?::([^@/?#]*))?@([^:/?#]+)(?::(\d+))?/([^?#]*)')
if (-not $am.Success -or -not $am.Groups[1].Value -or -not $am.Groups[2].Value) {
    throw 'DATABASE_URL tidak bisa diurai (format: postgres://user:pass@host:port/db).'
}
$appUser = $am.Groups[1].Value
$appPass = $am.Groups[2].Value
$pgHost = $sm.Groups[3].Value
$pgPort = if ($sm.Groups[4].Success) { $sm.Groups[4].Value } else { '5432' }
if (-not $sm.Groups[2].Value) { throw 'Password superuser kosong di DB_SUPERUSER_URL.' }
if ($appUser -notmatch '^[A-Za-z0-9_]+$') { throw "Nama role aplikasi tidak valid: $appUser." }
# Password role hanya boleh alfanumerik agar aman di-interpolasi ke SQL dan URL tanpa encoding.
if ($appPass -notmatch '^[A-Za-z0-9]{16,}$') { throw 'Password aplikasi di DATABASE_URL harus alfanumerik minimal 16 karakter.' }

$env:PGPASSWORD = $sm.Groups[2].Value
try {
    # Tes koneksi dulu dengan pesan yang jelas bila gagal.
    $null = & psql -h $pgHost -p $pgPort -U $sm.Groups[1].Value -d postgres -tAc 'SELECT 1'
    if ($LASTEXITCODE -ne 0) { throw 'Gagal konek ke PostgreSQL. Periksa host/port/user/password di DB_SUPERUSER_URL dan pastikan service PostgreSQL berjalan.' }

    # Role aplikasi: buat hanya bila belum ada; yang sudah ada tidak disentuh sama sekali.
    $roleFound = & psql -h $pgHost -p $pgPort -U $sm.Groups[1].Value -d postgres -tAc "SELECT 1 FROM pg_roles WHERE rolname='$appUser'"
    if ($LASTEXITCODE -ne 0) { throw "Gagal mengecek role $appUser." }
    if ("$roleFound".Trim() -eq '1') {
        Write-Host "[dbsetup] role $appUser sudah ada - password tidak diubah."
    } else {
        $null = & psql -h $pgHost -p $pgPort -U $sm.Groups[1].Value -d postgres -c "CREATE ROLE $appUser WITH LOGIN PASSWORD '$appPass'"
        if ($LASTEXITCODE -ne 0) { throw "Gagal membuat role $appUser." }
        Write-Host "[dbsetup] role $appUser dibuat."
    }

    foreach ($n in @($dbName, $testName)) {
        $found = & psql -h $pgHost -p $pgPort -U $sm.Groups[1].Value -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$n'"
        if ($LASTEXITCODE -ne 0) { throw "Gagal mengecek database $n." }
        if ("$found".Trim() -eq '1') {
            Write-Host "[dbsetup] database $n sudah ada - dilewati."
        } else {
            $null = & psql -h $pgHost -p $pgPort -U $sm.Groups[1].Value -d postgres -c "CREATE DATABASE $n OWNER $appUser ENCODING 'UTF8'"
            if ($LASTEXITCODE -ne 0) { throw "Gagal membuat database $n." }
            Write-Host "[dbsetup] database $n dibuat (owner $appUser)."
        }
    }
    Write-Host '[dbsetup] selesai. Tidak ada konfigurasi auth yang diubah.'
} finally {
    # Password wajib dibersihkan dari environment walau terjadi error di atas.
    Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
}
