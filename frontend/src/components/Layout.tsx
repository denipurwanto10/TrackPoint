import React from 'react';
import { Link, NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth';
import { applyTheme, saveTheme } from '../theme';
import {
  IconBox, IconDashboard, IconGrid, IconLayers, IconLogout, IconMoon, IconSearch, IconShield, IconSun, IconUsers, IconWrench, Logo,
} from './Icons';
import { HeadSlotContext, Toaster, initials, statusLabel, CommandPalette, CommandItem, NotificationDropdown, NotificationItem, ThemeToggle } from './ui';

interface Item {
  to: string;
  label: string;
  icon: React.FC<{ className?: string }>;
  end?: boolean;
  admin?: boolean;
}

const MAIN: Item[] = [
  { to: '/', label: 'Dashboard', icon: IconDashboard, end: true },
  { to: '/aset', label: 'Aset', icon: IconBox },
  { to: '/maintenance', label: 'Maintenance', icon: IconWrench },
];
const ADMIN: Item[] = [
  { to: '/master', label: 'Kategori & lokasi', icon: IconLayers, admin: true },
  { to: '/pengguna', label: 'Pengguna', icon: IconUsers, admin: true },
];
const ACCOUNT: Item[] = [{ to: '/sesi', label: 'Sesi saya', icon: IconShield }];

const TITLES: Record<string, string> = {
  '/': 'Dashboard', '/aset': 'Aset', '/maintenance': 'Maintenance',
  '/master': 'Kategori & lokasi', '/pengguna': 'Pengguna', '/sesi': 'Sesi saya',
};

const useTheme = () => {
  const [dark, setDark] = React.useState(() => document.documentElement.dataset.theme === 'dark');
  const toggle = () => {
    const next = !dark;
    setDark(next);
    applyTheme(next ? 'dark' : 'light');
    saveTheme(next ? 'dark' : 'light');
  };
  return { dark, toggle };
};

const Layout: React.FC = () => {
  const { dark, toggle: toggleTheme } = useTheme();
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const [open, setOpen] = React.useState(false);
  const [cmdOpen, setCmdOpen] = React.useState(false);
  const [slot, setSlot] = React.useState<HTMLElement | null>(null);

  React.useEffect(() => { setOpen(false); }, [pathname]);
  React.useEffect(() => { document.title = `${TITLES[pathname] ?? 'TrackPoint'} · TrackPoint`; }, [pathname]);
  React.useEffect(() => {
    if (!open) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setOpen(false); };
    document.addEventListener('keydown', onKey);
    return () => { document.body.style.overflow = prev; document.removeEventListener('keydown', onKey); };
  }, [open]);

  // Cmd+K global shortcut
  React.useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') { e.preventDefault(); setCmdOpen(true); }
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  const role = user?.role;
  const isAdmin = role === 'ADMIN';
  const all = [...MAIN, ...ADMIN, ...ACCOUNT].filter((i) => !i.admin || isAdmin);
  const doLogout = () => { logout().catch(() => navigate('/login')); };
  const themeLabel = dark ? 'Mode terang' : 'Mode gelap';

  // Item palet perintah. Pintasan "G lalu huruf" ditangani di efek di bawah.
  const cmdItems: CommandItem[] = React.useMemo(() => [
    { section: 'Halaman', label: 'Dashboard', description: 'Ringkasan aset dan aktivitas', shortcut: 'G D', icon: <IconDashboard className="h-[18px] w-[18px]" />, onSelect: () => navigate('/') },
    { section: 'Halaman', label: 'Aset', description: 'Kelola daftar aset', shortcut: 'G A', icon: <IconBox className="h-[18px] w-[18px]" />, onSelect: () => navigate('/aset') },
    { section: 'Halaman', label: 'Maintenance', description: 'Tiket kerusakan dan perbaikan', shortcut: 'G M', icon: <IconWrench className="h-[18px] w-[18px]" />, onSelect: () => navigate('/maintenance') },
    ...(isAdmin ? [
      { section: 'Halaman', label: 'Kategori & lokasi', description: 'Data master kategori dan lokasi', shortcut: 'G K', icon: <IconLayers className="h-[18px] w-[18px]" />, onSelect: () => navigate('/master') },
      { section: 'Halaman', label: 'Pengguna', description: 'Kelola akun pengguna', shortcut: 'G U', icon: <IconUsers className="h-[18px] w-[18px]" />, onSelect: () => navigate('/pengguna') },
    ] : []),
    { section: 'Halaman', label: 'Sesi saya', description: 'Lihat dan kelola sesi aktif', shortcut: 'G S', icon: <IconShield className="h-[18px] w-[18px]" />, onSelect: () => navigate('/sesi') },
    { section: 'Pengaturan', label: themeLabel, description: 'Ganti tema tampilan', icon: dark ? <IconSun className="h-[18px] w-[18px]" /> : <IconMoon className="h-[18px] w-[18px]" />, onSelect: toggleTheme },
  ], [dark, isAdmin, navigate, toggleTheme, themeLabel]);

  // "G" lalu huruf (dalam 1 detik) membuka halaman; diabaikan saat mengetik di kolom isian.
  React.useEffect(() => {
    const routes: Record<string, string> = { d: '/', a: '/aset', m: '/maintenance', s: '/sesi', ...(isAdmin ? { k: '/master', u: '/pengguna' } : {}) };
    let armed = 0;
    const onKey = (e: KeyboardEvent) => {
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      const t = e.target as HTMLElement | null;
      if (t && (t.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName))) return;
      const k = e.key.toLowerCase();
      if (k === 'g') { armed = Date.now(); return; }
      if (armed && Date.now() - armed < 1000 && routes[k]) { navigate(routes[k]); }
      armed = 0;
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, [isAdmin, navigate]);

  // Notifikasi: ganti dengan API nyata. Demo data hanya di development.
    const [notifs, setNotifs] = React.useState<NotificationItem[]>([]);

    React.useEffect(() => {
      const load = async () => {
        try {
          // API endpoint belum ada - gunakan data demo di development
          if (import.meta.env.DEV) {
            setNotifs([
              { id: '1', title: 'Aset baru ditambahkan', message: 'Laptop ThinkPad X1 Carbon (TP-LAP-0142) masuk inventaris', time: '5 menit lalu', unread: true },
              { id: '2', title: 'Tiket maintenance diselesaikan', message: 'Access Point Ubiquiti U6-LR (TP-NET-0018) sudah diperbaiki', time: '2 jam lalu', unread: true },
              { id: '3', title: 'Pemeliharaan terjadwal', message: 'Printer Canon imageRUNNER 2630 jadwal service besok', time: '1 hari lalu', unread: false },
            ]);
          }
        } finally {
        }
      };
      load();
    }, []);

    const markAllRead = React.useCallback(() => setNotifs((l) => l.map((n) => ({ ...n, unread: false }))), []);
    const notifItems = React.useMemo(
      () => notifs.map((n) => ({ ...n, onClick: () => setNotifs((l) => l.map((x) => (x.id === n.id ? { ...x, unread: false } : x))) })),
      [notifs],
    );
  const goTickets = React.useCallback(() => navigate('/maintenance'), [navigate]);

  const link = ({ to, label, icon: Icon, end }: Item) => (
    <NavLink
      key={to}
      to={to}
      end={end}
      title={label}
      className={({ isActive }) => `nav-link justify-center lg:justify-start ${isActive ? 'active' : ''}`}
    >
      <Icon className="h-[18px] w-[18px] shrink-0" />
      <span className="hidden lg:inline">{label}</span>
    </NavLink>
  );

  return (
    <div className="min-h-screen">
      <a href="#main" className="skip-link">Lewati ke konten</a>

      {/* ---------- Sidebar melayang: rail ikon di tablet, penuh di desktop ---------- */}
      <aside className="sidebar fixed bottom-4 left-4 top-4 z-30 hidden w-[4.5rem] flex-col overflow-y-auto px-2.5 py-5 md:flex lg:w-64 lg:px-4">
        <Link to="/" className="mb-7 flex items-center justify-center gap-2.5 rounded-lg lg:justify-start lg:px-2" aria-label="TrackPoint">
          <Logo size={34} />
          <span className="hidden text-xl lg:block" style={{ fontFamily: "var(--font-display)", letterSpacing: "-0.01em" }}>TrackPoint</span>
        </Link>

        <nav className="flex flex-1 flex-col" aria-label="Navigasi utama">
          <div className="nav-group">
            <p className="nav-label">Menu</p>
            <div className="flex flex-col gap-1">{MAIN.map(link)}</div>
          </div>
          {isAdmin && (
            <div className="nav-group">
              <p className="nav-label">Admin</p>
              <div className="flex flex-col gap-1">{ADMIN.map(link)}</div>
            </div>
          )}
          <div className="nav-group">
            <p className="nav-label">Akun</p>
            <div className="flex flex-col gap-1">
              {ACCOUNT.map(link)}
              <button onClick={doLogout} className="nav-link justify-center lg:justify-start" title="Keluar" aria-label="Keluar">
                <IconLogout className="h-[18px] w-[18px] shrink-0" />
                <span className="hidden lg:inline">Keluar</span>
              </button>
            </div>
          </div>

          <div className="promo mt-auto hidden lg:block">
            <p>Ada kendala aset?</p>
            <p>Laporkan kerusakan dan pantau penanganannya lewat tiket.</p>
            <Link to="/maintenance" className="btn btn-primary mt-3 w-full">Buka tiket</Link>
          </div>
        </nav>
      </aside>

      <div className="md:pl-[6.25rem] lg:pl-[18.25rem]">
        {/* ---------- Topbar: hanya di mobile ---------- */}
        <header className="topbar sticky top-0 z-20 md:hidden">
          <div className="flex h-14 items-center gap-2 px-4">
            <Link to="/" className="flex h-11 items-center gap-2.5" aria-label="TrackPoint">
              <Logo size={30} />
              <span className="text-base font-extrabold tracking-tight">TrackPoint</span>
            </Link>
            <button className="tool-btn ml-auto !h-10 !w-10" aria-label="Cari atau ketik perintah" title="Cari atau ketik perintah (⌘K)" onClick={() => setCmdOpen(true)}>
              <IconSearch className="h-5 w-5" />
            </button>
            <NotificationDropdown items={notifItems} onMarkAllRead={markAllRead} onViewAll={goTickets} />
            <ThemeToggle dark={dark} onToggle={toggleTheme} aria-label={themeLabel} />
          </div>
        </header>

        <main id="main" tabIndex={-1} className="mx-auto w-full max-w-[90rem] px-4 pb-28 pt-5 outline-none md:px-6 md:pb-12 md:pt-6 lg:px-8">
          {/* baris atas: judul halaman (kiri) + cari, notifikasi, tema, profil (kanan) */}
          <div className="head-row mb-6">
            <div ref={setSlot} className="head-slot" />
            <div className="head-tools hidden md:flex">
              <button
                className="tool-btn"
                aria-label="Cari atau ketik perintah"
                title="Cari atau ketik perintah (⌘K)"
                onClick={() => setCmdOpen(true)}
              >
                <IconSearch className="h-[18px] w-[18px]" />
              </button>
              <NotificationDropdown items={notifItems} onMarkAllRead={markAllRead} onViewAll={goTickets} />
              <ThemeToggle dark={dark} onToggle={toggleTheme} aria-label={themeLabel} />
              <Link to="/sesi" className="user-chip" title="Sesi saya">
                <span className="avatar">{initials(user?.nama)}</span>
                <span className="hidden min-w-0 text-left xl:block">
                  <span className="block max-w-[10rem] truncate text-[0.8125rem] font-bold leading-tight">{user?.nama}</span>
                  <span className="block text-xs leading-tight" style={{ color: 'var(--muted)' }}>{role ? statusLabel(role) : ''}</span>
                </span>
              </Link>
            </div>
          </div>

          <HeadSlotContext.Provider value={slot}>
            <div key={pathname} className="page-enter">
              <Outlet />
            </div>
          </HeadSlotContext.Provider>
        </main>
      </div>

      <Toaster />

      {/* ---------- Command Palette (global) ---------- */}
      <CommandPalette isOpen={cmdOpen} onClose={() => setCmdOpen(false)} items={cmdItems} />


      {/* ---------- Bottom tab bar (mobile) ---------- */}
      <nav className="bottomnav fixed inset-x-0 bottom-0 z-30 flex md:hidden" aria-label="Navigasi utama">
        {MAIN.map(({ to, label, icon: Icon, end }) => (
          <NavLink key={to} to={to} end={end} className={({ isActive }) => `tab ${isActive ? 'active' : ''}`}>
            <Icon className="h-[22px] w-[22px]" />
            {label}
          </NavLink>
        ))}
        <button className={`tab ${open ? 'active' : ''}`} onClick={() => setOpen(true)} aria-label="Menu lainnya" aria-haspopup="dialog" aria-expanded={open}>
          <IconGrid className="h-[22px] w-[22px]" />
          Lainnya
        </button>
      </nav>

      {/* ---------- Sheet menu (mobile) ---------- */}
      {open && (
        <div className="fixed inset-0 z-50 md:hidden" role="dialog" aria-modal="true" aria-label="Menu">
          <div className="modal-backdrop" onClick={() => setOpen(false)} />
          <div className="drawer-sheet">
            <div className="mx-auto mb-3 h-1 w-9 rounded-full" style={{ background: 'var(--line-strong)' }} />
            <div className="mb-2 flex items-center gap-3 px-3 py-2">
              <span className="avatar">{initials(user?.nama)}</span>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-bold">{user?.nama}</p>
                <p className="truncate text-xs" style={{ color: 'var(--muted)' }}>{user?.email}</p>
              </div>
            </div>
            <div className="space-y-1 border-t pt-2" style={{ borderColor: 'var(--line)' }}>
              {all.map(({ to, label, icon: Icon, end }) => (
                <NavLink key={to} to={to} end={end} className={({ isActive }) => `menu-link ${isActive ? 'active' : ''}`}>
                  <Icon className="h-5 w-5" /> {label}
                </NavLink>
              ))}
            </div>
            <button onClick={doLogout} className="btn btn-ghost mt-3 w-full" style={{ height: '2.75rem' }}>
              <IconLogout className="h-[18px] w-[18px]" /> Keluar
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default Layout;
