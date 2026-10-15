-- Core schema TrackPoint (goose). Idempotent per-version, append-only histori.
--
-- Hanya bagian Up yang disengaja. Tidak ada Down: mencegah reset destruktif
-- tak sengaja pada database pengembangan. Test DB dibangun/dibersihkan terpisah.

-- +goose Up

CREATE TYPE user_role AS ENUM ('ADMIN', 'TEKNISI', 'KARYAWAN');
CREATE TYPE asset_status AS ENUM ('AVAILABLE', 'ASSIGNED', 'IN_MAINTENANCE', 'INACTIVE');
CREATE TYPE maintenance_status AS ENUM ('REPORTED', 'IN_PROGRESS', 'DONE', 'CANCELLED');
CREATE TYPE maintenance_return_to AS ENUM ('AVAILABLE', 'ASSIGNED', 'INACTIVE');

-- Nomor aset: sequence Postgres, bukan MAX(id)+1 di aplikasi.
CREATE SEQUENCE asset_code_seq START 1;

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama          TEXT NOT NULL,
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'KARYAWAN',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));
CREATE INDEX users_role_active_idx ON users (role, is_active);

CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama       TEXT NOT NULL,
    deskripsi  TEXT,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX categories_nama_key ON categories (lower(nama));

CREATE TABLE locations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nama       TEXT NOT NULL,
    gedung     TEXT,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX locations_nama_key ON locations (lower(nama));

-- Kondisi terkini aset = cache dari histori terakhir; sumber kebenaran perpindahan = asset_movements.
CREATE TABLE assets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kode          TEXT NOT NULL UNIQUE,
    nama          TEXT NOT NULL,
    serial_number TEXT UNIQUE,
    category_id   UUID NOT NULL REFERENCES categories (id),
    location_id   UUID NOT NULL REFERENCES locations (id),
    holder_id     UUID REFERENCES users (id),
    status        asset_status NOT NULL DEFAULT 'AVAILABLE',
    purchase_date DATE,
    catatan       TEXT,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    -- Aset hanya boleh punya pemegang (holder_id) saat status = ASSIGNED.
    -- Boolean equality: (status = 'ASSIGNED') = (holder_id IS NOT NULL)
    -- Izinkan: AVAILABLE+NULL, ASSIGNED+NOT NULL, IN_MAINTENANCE+NULL, INACTIVE+NULL.
    -- Tolak: AVAILABLE+NOT NULL, ASSIGNED+NULL, IN_MAINTENANCE+NOT NULL, INACTIVE+NOT NULL.
    CONSTRAINT assets_status_holder_check CHECK (
        (status = 'ASSIGNED') = (holder_id IS NOT NULL)
    )
);
CREATE INDEX assets_status_idx ON assets (status);
CREATE INDEX assets_location_idx ON assets (location_id);
CREATE INDEX assets_holder_idx ON assets (holder_id);
CREATE INDEX assets_category_idx ON assets (category_id);

-- Append-only. Tidak ada UPDATE/DELETE oleh aplikasi (koreksi = record baru is_correction).
CREATE TABLE asset_movements (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id           UUID NOT NULL REFERENCES assets (id),
    from_location_id   UUID REFERENCES locations (id),
    to_location_id     UUID NOT NULL REFERENCES locations (id),
    from_holder_id     UUID REFERENCES users (id),
    to_holder_id       UUID REFERENCES users (id),
    catatan            TEXT,
    is_correction      BOOLEAN NOT NULL DEFAULT FALSE,
    correction_reason  TEXT,
    created_by         UUID NOT NULL REFERENCES users (id),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT asset_movements_correction_check CHECK (
        (NOT is_correction) OR correction_reason IS NOT NULL
    )
);
CREATE INDEX asset_movements_asset_created_idx ON asset_movements (asset_id, created_at DESC);
CREATE INDEX asset_movements_created_idx ON asset_movements (created_at DESC);

CREATE TABLE maintenance_records (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asset_id       UUID NOT NULL REFERENCES assets (id),
    asset_code     TEXT NOT NULL,
    status         maintenance_status NOT NULL DEFAULT 'REPORTED',
    keluhan        TEXT NOT NULL,
    tindakan       TEXT,
    biaya          NUMERIC(14, 2),
    return_to      maintenance_return_to,
    reported_by    UUID NOT NULL REFERENCES users (id),
    technician_id  UUID REFERENCES users (id),
    cancel_reason  TEXT,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT maintenance_done_check CHECK (
        status <> 'DONE' OR (tindakan IS NOT NULL AND return_to IS NOT NULL AND completed_at IS NOT NULL)
    ),
    CONSTRAINT maintenance_cancel_check CHECK (
        status <> 'CANCELLED' OR cancel_reason IS NOT NULL
    )
);
-- Aturan bisnis: satu aset hanya boleh punya satu tiket terbuka.
CREATE UNIQUE INDEX maintenance_one_open_ticket_idx
    ON maintenance_records (asset_id)
    WHERE status IN ('REPORTED', 'IN_PROGRESS');
CREATE INDEX maintenance_asset_created_idx ON maintenance_records (asset_id, created_at DESC);
CREATE INDEX maintenance_status_idx ON maintenance_records (status);

-- Refresh session: hanya hash token yang disimpan. Rotation + reuse detection.
CREATE TABLE refresh_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users (id),
    token_hash  TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    replaced_by UUID REFERENCES refresh_sessions (id),
    user_agent  TEXT,
    ip_address  TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT refresh_sessions_expiry_check CHECK (expires_at > created_at)
);
CREATE INDEX refresh_sessions_user_idx ON refresh_sessions (user_id);
CREATE INDEX refresh_sessions_active_idx ON refresh_sessions (expires_at) WHERE revoked_at IS NULL;