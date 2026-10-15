import React, { useState } from 'react';
import { useAuth } from './auth';
import type { ApiError } from './api';
import { IconArrowRight, IconBox, IconEye, IconEyeOff, IconLock, IconMail, IconMoon, IconShield, IconSun, IconWrench, Logo } from './components/Icons';
import { applyTheme, saveTheme, storedTheme, systemTheme } from './theme';
import { ErrorAlert } from './components/ui';

const FEATURES = [
  { icon: <IconBox className="h-[18px] w-[18px]" />, t: 'Inventaris terpusat', d: 'Setiap aset punya kode, lokasi, dan pemegang yang jelas.' },
  { icon: <IconWrench className="h-[18px] w-[18px]" />, t: 'Tiket maintenance', d: 'Lapor, kerjakan, dan tutup perbaikan dalam satu alur.' },
  { icon: <IconShield className="h-[18px] w-[18px]" />, t: 'Akses berbasis peran', d: 'Admin, teknisi, dan karyawan hanya melihat yang relevan.' },
];

const Login: React.FC = () => {
  const { login } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPw, setShowPw] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [caps, setCaps] = useState(false);
  const [dark, setDark] = useState(() => (storedTheme() ?? systemTheme()) === 'dark');

  const toggleTheme = () => {
    const next = !dark;
    setDark(next);
    applyTheme(next ? 'dark' : 'light');
    saveTheme(next ? 'dark' : 'light');
  };
  const onCaps = (e: React.KeyboardEvent) => setCaps(e.getModifierState?.('CapsLock') ?? false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await login(email, password);
    } catch (err) {
      // Pesan dari API dipakai apa adanya supaya kasus seperti "Akun dinonaktifkan"
      // tidak tertutup oleh pesan generik yang menyesatkan.
      const msg = (err as Partial<ApiError>)?.message;
      setError(msg && msg.trim() ? msg : 'Login gagal. Periksa email atau password.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="lg-root">
      {/* panel brand — desktop & tablet landscape */}
      <aside className="lg-art">
        <span className="lg-orb a" aria-hidden="true" />
        <span className="lg-orb b" aria-hidden="true" />
        <div className="lg-brand">
          <Logo size={36} />
          <span>TrackPoint</span>
        </div>

        <div>
          <h2 className="lg-headline">
            Setiap aset,<br /><em>selalu terlacak.</em>
          </h2>
          <p className="lg-lead">Platform manajemen aset IT dan tiket maintenance untuk tim yang menuntut kejelasan di setiap perangkat.</p>

          <div className="lg-stats" aria-hidden="true">
            <div className="lg-glass lg-stat"><b>100%</b><span>Aset tercatat</span></div>
            <div className="lg-glass lg-stat"><b>24/7</b><span>Pantau status</span></div>
            <div className="lg-glass lg-stat"><b>3</b><span>Peran akses</span></div>
          </div>

          <ul className="lg-feats">
            {FEATURES.map(({ icon, t, d }) => (
              <li key={t} className="lg-feat">
                <i>{icon}</i>
                <div>
                  <p>{t}</p>
                  <p>{d}</p>
                </div>
              </li>
            ))}
          </ul>
        </div>

        <p className="lg-foot">© {new Date().getFullYear()} TrackPoint</p>
      </aside>

      {/* form */}
      <main className="lg-main">
        <div className="lg-top">
          <div className="lg-brand">
            <Logo size={32} />
            <span>TrackPoint</span>
          </div>
          <button type="button" className="lg-theme" onClick={toggleTheme} aria-label={dark ? 'Ganti ke tema terang' : 'Ganti ke tema gelap'}>
            {dark ? <IconSun className="h-[18px] w-[18px]" /> : <IconMoon className="h-[18px] w-[18px]" />}
          </button>
        </div>

        <div className="lg-card">
          <h1 className="lg-title">Selamat datang</h1>
          <p className="lg-sub">Masuk untuk mengelola aset IT dan tiket maintenance.</p>

          <form onSubmit={handleSubmit} className="lg-form">
            {error && <ErrorAlert message={error} />}
            <div>
              <label className="label" htmlFor="email">Email</label>
              <div className="lg-field">
                <IconMail className="lg-ico" />
                <input
                  id="email"
                  type="email"
                  required
                  autoComplete="email"
                  inputMode="email"
                  placeholder="nama@perusahaan.com"
                  className="input"
                  aria-invalid={error ? true : undefined}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
            </div>
            <div>
              <label className="label" htmlFor="password">Password</label>
              <div className="lg-field">
                <IconLock className="lg-ico" />
                <input
                  id="password"
                  type={showPw ? 'text' : 'password'}
                  required
                  autoComplete="current-password"
                  placeholder="Masukkan password"
                  className="input has-toggle"
                  aria-invalid={error ? true : undefined}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  onKeyUp={onCaps}
                  onKeyDown={onCaps}
                  onBlur={() => setCaps(false)}
                />
                <button
                  type="button"
                  onClick={() => setShowPw(!showPw)}
                  className="lg-pw-toggle"
                  aria-label={showPw ? 'Sembunyikan password' : 'Tampilkan password'}
                  aria-pressed={showPw}
                >
                  {showPw ? <IconEyeOff className="h-[18px] w-[18px]" /> : <IconEye className="h-[18px] w-[18px]" />}
                </button>
              </div>
              {caps && <p className="lg-hint" role="status">Caps Lock sedang aktif</p>}
            </div>
            <button type="submit" disabled={loading} className="lg-submit">
              {loading ? <span className="spinner" aria-hidden="true" /> : null}
              <span>{loading ? 'Memproses…' : 'Masuk'}</span>
              {!loading && <IconArrowRight />}
            </button>
          </form>

          <p className="lg-secure"><IconShield />Koneksi aman · akses sesuai peran Anda</p>
        </div>
      </main>
    </div>
  );
};

export default Login;
