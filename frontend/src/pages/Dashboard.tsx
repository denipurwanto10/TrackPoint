import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { apiFetch } from '../api';
import { useAuth } from '../auth';
import {
  IconBox, IconChevronRight, IconPlus, IconShield, IconSwap, IconTicket, IconUser, IconWrench,
} from '../components/Icons';
import { Empty, ErrorAlert, PageHeader, StatusBadge } from '../components/ui';

interface Dash {
  total_aset: number;
  tersedia: number;
  dipakai: number;
  maintenance: number;
  nonaktif: number;
  tiket_terbuka: number;
  aset_saya: number;
  tiket_saya_terbuka: number;
  mutasi_terbaru: Array<{ id: string; asset_kode: string; asset_nama: string; to_location_nama: string; created_at: string }>;
  tiket_perlu_ditangani: Array<{ id: string; asset_code: string; asset_nama: string; keluhan: string; status: string }>;
}

/** Hormati setelan sistem: kalau animasi dimatikan, angka/kurva langsung tampil final. */
const reduceMotion = () => window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false;

const greeting = (hour: number) => {
  if (hour < 11) return 'Selamat pagi';
  if (hour < 15) return 'Selamat siang';
  if (hour < 18) return 'Selamat sore';
  return 'Selamat malam';
};

const roleLabel = (role?: string) => {
  if (role === 'ADMIN') return 'Admin';
  if (role === 'TEKNISI') return 'Teknisi';
  if (role === 'KARYAWAN') return 'Karyawan';
  return 'Pengguna';
};

const fmtDate = (s: string) =>
  new Date(s).toLocaleString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' });

/** Bilangan berjalan naik saat kartu pertama muncul. */
const useCountUp = (target: number, on: boolean) => {
  const [v, setV] = useState(0);
  useEffect(() => {
    if (!on) return;
    if (reduceMotion()) { setV(target); return; }
    let raf = 0;
    const t0 = performance.now();
    const tick = (t: number) => {
      const p = Math.min(1, (t - t0) / 650);
      setV(Math.round(target * (1 - Math.pow(1 - p, 3))));
      if (p < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [target, on]);
  return v;
};

/** Jam berjalan di panel hero — dijadwalkan ulang tiap detik, dibersihkan saat unmount. */
const useClock = () => {
  const [now, setNow] = useState(() => new Date());
  useEffect(() => {
    const id = window.setInterval(() => setNow(new Date()), 1000);
    return () => window.clearInterval(id);
  }, []);
  return now;
};

/** `c` = nada terang untuk di atas panel hero, `cc` = nada pekat untuk di atas kartu. */
const statusOf = (d: Dash) => [
  { k: 'Tersedia', v: d.tersedia, c: 'var(--seg-1)', cc: 'var(--seg-1-card)', pill: 'pill-ok' },
  { k: 'Dipakai', v: d.dipakai, c: 'var(--seg-2)', cc: 'var(--seg-2-card)', pill: 'pill-info' },
  { k: 'Maintenance', v: d.maintenance, c: 'var(--seg-3)', cc: 'var(--seg-3-card)', pill: 'pill-warn' },
  { k: 'Nonaktif', v: d.nonaktif, c: 'var(--seg-4)', cc: 'var(--seg-4-card)', pill: 'pill-mute' },
];

// skala sumbu Y: langkah bulat (1/2/5/10/20/…) + headroom ≥8% agar batang tertinggi tidak menempel garis atas
const NICE_STEPS = [1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000];
const niceMax = (n: number) => (NICE_STEPS.find((s) => s * 4 >= Math.max(4, n * 1.08)) ?? 1000) * 4;

const percent = (v: number, total: number) => (total > 0 ? Math.round((v / total) * 100) : 0);

/* ================= panel hero ================= */

const Hero: React.FC<{ name: string; role?: string; d: Dash | null; grown: boolean }> = ({ name, role, d, grown }) => {
  const now = useClock();
  const segs = d ? statusOf(d) : [];
  const total = d?.total_aset ?? 0;

  return (
    <section className="dash-hero" aria-labelledby="hero-title">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
        <div className="min-w-0">
          <span className="hero-kicker"><IconShield className="h-3.5 w-3.5" />{roleLabel(role)}</span>
          <h1 id="hero-title" className="hero-title mt-3">{greeting(now.getHours())}, {name}</h1>
          <p className="hero-meta">
            <span>{now.toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</span>
            <span className="sep" aria-hidden="true" />
            <time className="hero-clock" dateTime={now.toISOString()}>
              {now.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' })}
            </time>
            {d && (
              <>
                <span className="sep" aria-hidden="true" />
                <span>{d.tersedia} dari {d.total_aset} aset siap dipakai</span>
              </>
            )}
          </p>
        </div>
        <div className="hero-actions lg:pt-1.5">
          <Link to="/aset" className="btn btn-white"><IconPlus className="h-[18px] w-[18px]" />Tambah aset</Link>
          <Link to="/maintenance" className="btn btn-onhero"><IconTicket className="h-[18px] w-[18px]" />Tiket maintenance</Link>
        </div>
      </div>

      <div className="mt-6">
        <div className="flex items-baseline justify-between gap-3 text-xs font-semibold text-white/80">
          <span>Sebaran aset</span>
          {d && <span className="tabular-nums">{total} aset</span>}
        </div>
        {d ? (
          <div className="rail mt-2" role="img" aria-label={segs.map((s) => `${s.k} ${s.v}`).join(', ')}>
            {segs.map((s) => (
              <span
                key={s.k}
                className="rail-seg"
                title={`${s.k}: ${s.v} aset`}
                style={{ ['--c' as string]: s.c, width: grown ? `${(s.v / (total || 1)) * 100}%` : '0%' }}
              />
            ))}
          </div>
        ) : (
          <div className="rail mt-2" aria-hidden="true" />
        )}
        {d && (
          <ul className="rail-legend">
            {segs.map((s) => (
              <li key={s.k} style={{ ['--c' as string]: s.c }}>
                <i aria-hidden="true" />{s.k}<b>{s.v}</b>
                <span className="opacity-75">{percent(s.v, total)}%</span>
              </li>
            ))}
          </ul>
        )}
      </div>
    </section>
  );
};

/* ================= kartu KPI ================= */

const Tile: React.FC<{
  to: string; hero?: boolean; k: string; icon: React.FC<{ className?: string }>;
  label: string; value: number; note: string; pill: string; pillCls: string;
  grown: boolean; meter?: { pct: number; left: string; right: string };
}> = ({ to, hero, k, icon: Icon, label, value, note, pill, pillCls, grown, meter }) => {
  const n = useCountUp(value, grown);
  return (
    <Link to={to} className={`stat ${hero ? 'stat-hero' : ''}`} style={{ ['--k' as string]: k }}>
      <span className="stat-icon"><Icon className="h-5 w-5" /></span>
      <span className={`pill ${pillCls}`}>{pill}</span>
      <span className="stat-label">{label}</span>
      <span className="stat-num">{n}</span>
      <span className="stat-note">{note}</span>
      {meter && (
        <span className="stat-meter">
          <span className="meter" aria-hidden="true"><i style={{ width: grown ? `${meter.pct}%` : '0%' }} /></span>
          <span className="meter-cap"><span>{meter.left}</span><span>{meter.right}</span></span>
        </span>
      )}
    </Link>
  );
};

const CardHead: React.FC<{ id?: string; title: string; sub?: string; to?: string; cta?: string; right?: React.ReactNode }> = ({ id, title, sub, to, cta, right }) => (
  <div className="card-head">
    <div className="min-w-0">
      <h2 id={id} className="section-title">{title}</h2>
      {sub && <p className="card-sub">{sub}</p>}
    </div>
    {right}
    {to && (
      <Link to={to} className="link flex shrink-0 items-center gap-0.5 text-[0.8125rem]">
        {cta} <IconChevronRight className="h-3.5 w-3.5" />
      </Link>
    )}
  </div>
);

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [d, setD] = useState<Dash | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [grown, setGrown] = useState(false);
  const [hover, setHover] = useState<number | null>(null);

  useEffect(() => {
    apiFetch('/api/v1/dashboard').then(setD).catch((e) => setErr(e.message ?? 'Gagal memuat dashboard.'));
  }, []);

  // Gerak otomatis satu-satunya: cincin, batang, dan bilah tumbuh saat data tiba.
  useEffect(() => {
    if (!d) return;
    if (reduceMotion()) { setGrown(true); return; }
    const id = requestAnimationFrame(() => requestAnimationFrame(() => setGrown(true)));
    return () => cancelAnimationFrame(id);
  }, [d]);

  const name = user?.nama?.split(' ')[0] ?? 'Pengguna';
  const head = <PageHeader title="Dashboard" />;
  const hero = <Hero name={name} role={user?.role} d={d} grown={grown} />;

  if (err) {
    return (
      <>
        {head}
        {hero}
        <div className="mt-5"><ErrorAlert message={err} /></div>
      </>
    );
  }

  if (!d) {
    return (
      <>
        {head}
        <div aria-busy="true" aria-label="Memuat">
          <div className="skeleton h-[15.5rem] !rounded-3xl" />
          <div className="mt-5 grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
            {[0, 1, 2, 3].map((i) => <div key={i} className="skeleton h-[8.75rem] !rounded-3xl" />)}
          </div>
          <div className="mt-5 grid gap-5 md:grid-cols-2 xl:grid-cols-12">
            <div className="skeleton h-80 !rounded-3xl xl:col-span-7" />
            <div className="skeleton h-80 !rounded-3xl xl:col-span-5" />
            <div className="skeleton h-72 !rounded-3xl xl:col-span-7" />
            <div className="skeleton h-72 !rounded-3xl xl:col-span-5" />
          </div>
        </div>
      </>
    );
  }

  const segs = statusOf(d);
  const ringR = [80, 63, 46, 29];
  const top = Math.max(...segs.map((s) => s.v));
  const topIdx = segs.findIndex((s) => s.v === top);
  const active = hover ?? topIdx;
  const yMax = niceMax(top);
  const pctTersedia = percent(d.tersedia, d.total_aset);

  return (
    <>
      {head}
      {hero}

      {/* ================= ringkasan ================= */}
      <section aria-label="Ringkasan" className="mt-5 grid auto-rows-fr grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
        <Tile to="/aset" hero k="var(--brand)" grown={grown} icon={IconBox} label="Total aset" value={d.total_aset}
          note="Aset terdaftar di sistem" pill={`${pctTersedia}% tersedia`} pillCls="pill-ok" />
        <Tile to="/maintenance" k="var(--dot-danger)" grown={grown} icon={IconTicket} label="Tiket terbuka" value={d.tiket_terbuka}
          note="Perlu ditangani" pill={d.tiket_terbuka > 0 ? 'Perlu aksi' : 'Aman'} pillCls={d.tiket_terbuka > 0 ? 'pill-danger' : 'pill-ok'} />
        <Tile to="/aset" k="var(--dot-info)" grown={grown} icon={IconUser} label="Aset yang Anda pegang" value={d.aset_saya}
          note="Sedang dalam tanggung jawab Anda" pill="Milik saya" pillCls="pill-info" />
        <Tile to="/maintenance" k="var(--dot-warn)" grown={grown} icon={IconWrench} label="Tiket Anda yang terbuka" value={d.tiket_saya_terbuka}
          note="Laporan yang belum selesai" pill={d.tiket_saya_terbuka > 0 ? 'Berjalan' : 'Selesai'} pillCls={d.tiket_saya_terbuka > 0 ? 'pill-warn' : 'pill-ok'} />
      </section>

      <div className="mt-5 grid gap-5 md:grid-cols-2 xl:grid-cols-12">
        {/* ================= kolom kiri ================= */}
        <section className="card flex min-w-0 flex-col xl:col-span-7" aria-labelledby="grafik">
          <div className="card-head">
            <div>
              <h2 id="grafik" className="section-title">Aset per status</h2>
              <p className="card-sub">Sebaran jumlah aset berdasarkan kondisinya</p>
            </div>
          </div>
          <div className="chart" role="img" aria-label={segs.map((s) => `${s.k} ${s.v}`).join(', ')}>
            <div className="chart-y" aria-hidden="true">
              {[0, 1, 2, 3, 4].map((i) => <span key={i}>{(yMax / 4) * i}</span>)}
            </div>
            <div className="chart-plot">
              {segs.map((s, i) => {
                const h = grown ? (s.v / yMax) * 100 : 0;
                return (
                  <button
                    key={s.k}
                    type="button"
                    className={`chart-col ${active === i ? 'on' : ''}`}
                    style={{ ['--c' as string]: s.cc }}
                    aria-label={`${s.k}: ${s.v} aset`}
                    onMouseEnter={() => setHover(i)}
                    onMouseLeave={() => setHover(null)}
                    onFocus={() => setHover(i)}
                    onBlur={() => setHover(null)}
                  >
                    <span className="chart-val" style={{ bottom: `${Math.min(h + 6, 84)}%` }}>{s.v}</span>
                    <span className="chart-bar" style={{ height: `${h}%` }} />
                  </button>
                );
              })}
            </div>
            <div className="chart-x" aria-hidden="true">
              {segs.map((s) => <span key={s.k}>{s.k}</span>)}
            </div>
          </div>
        </section>

        <section className="card min-w-0 xl:col-span-5" aria-labelledby="statistik">
          <div className="card-head">
            <div>
              <h2 id="statistik" className="section-title">Kesehatan aset</h2>
              <p className="card-sub">Porsi tiap status dari seluruh aset</p>
            </div>
          </div>

          <div className="relative">
            <svg viewBox="0 0 200 200" className="rings" role="img" aria-label={segs.map((s) => `${s.k} ${s.v}`).join(', ')}>
              <g transform="rotate(-90 100 100)">
                {segs.map((s, i) => {
                  const C = 2 * Math.PI * ringR[i];
                  const len = grown && d.total_aset > 0 ? (s.v / d.total_aset) * C : 0;
                  return (
                    <g key={s.k}>
                      <circle className="ring-track" cx="100" cy="100" r={ringR[i]} />
                      {s.v > 0 && (
                        <circle className="ring-arc" cx="100" cy="100" r={ringR[i]} stroke={s.cc} strokeDasharray={`${len} ${C}`} />
                      )}
                    </g>
                  );
                })}
              </g>
              <text className="ring-center" x="100" y="104" textAnchor="middle">{d.total_aset}</text>
              <text className="ring-center-sub" x="100" y="118" textAnchor="middle">Total aset</text>
            </svg>
          </div>

          <ul className="mt-4">
            {segs.map((s) => (
              <li key={s.k} className="stat-row">
                <span className="stat-row-dot" style={{ ['--c' as string]: s.cc }} />
                <span className="flex-1 text-sm font-semibold">{s.k}</span>
                <span className="num text-sm font-bold">{s.v}</span>
                <span className={`pill ${s.pill} num`}>{percent(s.v, d.total_aset)}%</span>
              </li>
            ))}
          </ul>
        </section>

        {/* ================= aktivitas ================= */}
        <section className="card min-w-0 xl:col-span-7" aria-labelledby="mutasi">
          <CardHead id="mutasi" title="Mutasi terbaru" sub="Perpindahan aset antar lokasi" to="/aset" cta="Semua aset" />
          {d.mutasi_terbaru.length === 0 && <Empty text="Belum ada mutasi" hint="Perpindahan aset antar lokasi akan muncul di sini." />}
          <ul>
            {d.mutasi_terbaru.map((m) => (
              <li key={m.id} className="feed-item">
                <span className="feed-icon"><IconSwap className="h-[18px] w-[18px]" /></span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold">{m.asset_nama}</p>
                  <p className="meta truncate">
                    <span className="mono">{m.asset_kode}</span> dipindah ke {m.to_location_nama}
                  </p>
                </div>
                <time className="meta num shrink-0" dateTime={m.created_at}>{fmtDate(m.created_at)}</time>
              </li>
            ))}
          </ul>
        </section>

        <section className="card min-w-0 xl:col-span-5" aria-labelledby="tiket">
          <CardHead id="tiket" title="Tiket perlu ditangani" sub="Laporan kerusakan terbaru" to="/maintenance" cta="Semua" />
          {d.tiket_perlu_ditangani.length === 0 && <Empty text="Tidak ada tiket terbuka" hint="Laporan kerusakan baru akan muncul di sini." />}
          <ul>
            {d.tiket_perlu_ditangani.map((t) => (
              <li key={t.id} className="feed-item">
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold">
                    <Link to="/maintenance" className="mono link mr-2">{t.asset_code}</Link>
                    {t.asset_nama}
                  </p>
                  <p className="meta truncate">{t.keluhan}</p>
                </div>
                <StatusBadge value={t.status} />
              </li>
            ))}
          </ul>
        </section>
      </div>
    </>
  );
};

export default Dashboard;