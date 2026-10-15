#!/usr/bin/env pwsh
# Setup lokal TrackPoint. Aman dijalankan ulang (idempoten).
# Tidak mengubah konfigurasi Windows, tidak menyentuh pg_hba, tidak memakai auth trust.
# Pakai:  pwsh -File scripts/setup.ps1  (jalan juga di powershell.exe 5.1)
# Encoding: UTF-8 with BOM + CRLF agar terbaca benar di PowerShell 5.1.

$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root '.env'
$example = Join-Path $root '.env.example'

function New-LocalSecret {
    param([int]$Bytes = 48)
    $buf = New-Object 'byte[]' $Bytes
    # RNGCryptoServiceProvider ada di .NET Framework (PS 5.1) maupun .NET modern.
    # (RandomNumberGenerator::Fill statis hanya ada di .NET Core, jadi tidak dipakai.)
    $rng = New-Object Security.Cryptography.RNGCryptoServiceProvider
    try {
        $rng.GetBytes($buf)
    } finally {
        $rng.Dispose()
    }
    [Convert]::ToBase64String($buf)
}

if (-not (Test-Path $envFile)) {
    if (-not (Test-Path $example)) { throw "Tidak ditemukan: $example" }
    Copy-Item $example $envFile
    Write-Host '[setup] .env dibuat dari .env.example'
} else {
    Write-Host '[setup] .env sudah ada - tidak ditimpa'
}

# Isi hanya placeholder; kredensial DB tetap kamu yang tentukan.
$content = Get-Content $envFile -Raw
$content = $content -replace 'JWT_ACCESS_SECRET=GANTI_DENGAN_BASE64_ACAK_MINIMAL_48_BYTE', "JWT_ACCESS_SECRET=$(New-LocalSecret)"
if ($content -match 'JWT_ACCESS_SECRET=GANTI') {
    $content = $content -replace 'JWT_ACCESS_SECRET=.*', "JWT_ACCESS_SECRET=$(New-LocalSecret)"
}
# Normalisasi bila URL masih memakai placeholder lama *** (aman diulang; no-op bila tidak cocok)
$content = $content -replace '(postgres://trackpoint:)\*\*\*(@)', '${1}GANTI_PASSWORD${2}'
$content = $content -replace '(postgres://postgres:)\*\*\*(@)', '${1}GANTI_PASSWORD${2}'
Set-Content -Path $envFile -Value $content -NoNewline
Write-Host '[setup] secret JWT di-generate lokal. Ganti password DB di .env bila sudah tahu.'

Write-Host ''
Write-Host 'Langkah berikutnya (JALANKAN SENDIRI, tidak dilakukan script ini):'
Write-Host '  1. isi DB_PASSWORD dan password superuser di .env'
Write-Host '  2. pwsh -File scripts/dbsetup.ps1  (buat role + DB trackpoint dan trackpoint_test, idempoten)'
Write-Host '  3. cd backend; go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres $env:DATABASE_URL up  (migration goose ke DB dev)'
Write-Host '  4. cd backend; go run ./cmd/api  (server backend di http://localhost:8080)'
Write-Host ''
Write-Host 'Secret JWT tidak pernah dicetak. Kalau perlu regenerate: hapus nilainya dari .env lalu jalankan ulang script ini.'
