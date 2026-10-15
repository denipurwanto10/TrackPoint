import React, { useEffect, useMemo, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import {
  IconAlert, IconBell, IconCheck, IconCheckPlain, IconChevronDown, IconChevronLeft, IconChevronRight,
  IconInbox, IconInfo, IconMoon, IconSearch, IconSun, IconX,
} from './Icons';

/* ---------- label status ---------- */
// Nilai dari API tetap enum; yang berubah hanya teks yang dilihat pengguna.
const STATUS: Record<string, { label: string; cls: string }> = {
  AVAILABLE: { label: 'Tersedia', cls: 'badge badge-ok' },
  ASSIGNED: { label: 'Dipakai', cls: 'badge badge-info' },
  IN_MAINTENANCE: { label: 'Maintenance', cls: 'badge badge-warn' },
  INACTIVE: { label: 'Nonaktif', cls: 'badge badge-mute' },
  REPORTED: { label: 'Dilaporkan', cls: 'badge badge-warn' },
  IN_PROGRESS: { label: 'Dikerjakan', cls: 'badge badge-info' },
  DONE: { label: 'Selesai', cls: 'badge badge-ok' },
  CANCELLED: { label: 'Dibatalkan', cls: 'badge badge-mute' },
  ADMIN: { label: 'Admin', cls: 'badge badge-brand' },
  TEKNISI: { label: 'Teknisi', cls: 'badge badge-info' },
  KARYAWAN: { label: 'Karyawan', cls: 'badge badge-mute' },
};

export const statusLabel = (value: string) =>
  STATUS[value]?.label ?? (value.charAt(0) + value.slice(1).toLowerCase()).replace(/_/g, ' ');

export const StatusBadge: React.FC<{ value: string }> = ({ value }) => (
  <span className={STATUS[value]?.cls ?? 'badge badge-mute'}>{statusLabel(value)}</span>
);

/* ---------- page header ---------- */
// Layout menyediakan slot di baris atas (di kiri tombol cari, notifikasi, dan profil).
// PageHeader dirender ke slot itu lewat portal, jadi semua halaman otomatis mengikuti.
export const HeadSlotContext = React.createContext<HTMLElement | null>(null);

export const PageHeader: React.FC<{ title: string; subtitle?: string; action?: React.ReactNode }> = ({ title, subtitle, action }) => {
  const slot = React.useContext(HeadSlotContext);
  if (!slot) return null;
  return createPortal(
    <>
      <div className="head-title min-w-0">
        <h1 className="page-title">{title}</h1>
        {subtitle && <p className="page-sub">{subtitle}</p>}
      </div>
      {action && <div className="head-action shrink-0">{action}</div>}
    </>,
    slot
  );
};

/* ---------- pesan ---------- */
export const ErrorAlert: React.FC<{ message: string }> = ({ message }) => (
  <div className="alert" role="alert">
    <IconAlert className="mt-0.5 h-[18px] w-[18px] shrink-0" />
    <span>{message}</span>
  </div>
);

export const Empty: React.FC<{ text: string; hint?: string; action?: React.ReactNode }> = ({ text, hint, action }) => (
  <div className="empty">
    <IconInbox className="h-6 w-6" />
    <p className="empty-title">{text}</p>
    {hint && <p className="empty-hint">{hint}</p>}
    {action && <div className="mt-3">{action}</div>}
  </div>
);

/* ---------- loading ---------- */
export const TableSkeleton: React.FC<{ rows?: number }> = ({ rows = 5 }) => (
  <div className="border-t" style={{ borderColor: 'var(--line)' }} aria-busy="true" aria-label="Memuat data">
    {Array.from({ length: rows }).map((_, i) => (
      <div key={i} className="flex items-center gap-4 border-b py-4" style={{ borderColor: 'var(--line)' }}>
        <div className="skeleton h-4 flex-[2]" />
        <div className="skeleton hidden h-4 flex-1 sm:block" />
        <div className="skeleton hidden h-4 flex-1 lg:block" />
        <div className="skeleton h-4 w-20" />
      </div>
    ))}
  </div>
);

/* ---------- toast ---------- */
export type ToastKind = 'ok' | 'err' | 'info';
type ToastItem = { id: number; text: string; kind: ToastKind; ttl: number };
const toastListeners = new Set<(t: ToastItem) => void>();
let toastSeq = 0;

/** Konfirmasi singkat setelah aksi. Panggil dari mana saja: toast('Aset ditambahkan') / toast('Gagal', 'err'). */
export const toast = (text: string, kind: ToastKind = 'ok') =>
  toastListeners.forEach((l) => l({ id: ++toastSeq, text, kind, ttl: kind === 'err' ? 5200 : 3600 }));

const TOAST_ICON: Record<ToastKind, React.FC<{ className?: string }>> = { ok: IconCheck, err: IconAlert, info: IconInfo };

// Satu kartu toast: timer-nya sendiri supaya bisa dijeda saat kursor/fokus berhenti di atasnya.
const ToastCard: React.FC<{ t: ToastItem; paused: boolean; onDone: (id: number) => void }> = ({ t, paused, onDone }) => {
  const [left, setLeft] = useState(t.ttl);
  useEffect(() => {
    if (paused) return;
    const id = window.setInterval(() => setLeft((v) => Math.max(0, v - 100)), 100);
    return () => clearInterval(id);
  }, [paused]);
  useEffect(() => { if (left <= 0) onDone(t.id); }, [left, t.id, onDone]);
  const Icon = TOAST_ICON[t.kind];
  return (
    <div className="toast" data-kind={t.kind}>
      <span className="toast-icon"><Icon className="h-[18px] w-[18px]" /></span>
      <p className="toast-text">{t.text}</p>
      <button type="button" className="toast-x" aria-label="Tutup notifikasi" onClick={() => onDone(t.id)}>
        <IconX className="h-4 w-4" />
      </button>
      <span className="toast-bar" style={{ transform: `scaleX(${left / t.ttl})` }} aria-hidden="true" />
    </div>
  );
};

const ToastViewport: React.FC = () => {
  const [items, setItems] = useState<ToastItem[]>([]);
  const [paused, setPaused] = useState(false);
  useEffect(() => {
    const on = (t: ToastItem) => setItems((p) => [...p.slice(-2), t]);
    toastListeners.add(on);
    return () => { toastListeners.delete(on); };
  }, []);
  const done = React.useCallback((id: number) => setItems((p) => p.filter((x) => x.id !== id)), []);
  return (
    <div
      className="toaster" role="status"
      aria-live={items.some((t) => t.kind === 'err') ? 'assertive' : 'polite'}
      onMouseEnter={() => setPaused(true)} onMouseLeave={() => setPaused(false)}
      onFocus={() => setPaused(true)} onBlur={() => setPaused(false)}
    >
      {items.map((t) => <ToastCard key={t.id} t={t} paused={paused} onDone={done} />)}
    </div>
  );
};

/* ---------- konfirmasi bertema (pengganti window.confirm) ---------- */
export type ConfirmOptions = {
  title: string;
  body?: string;
  confirmLabel?: string;
  icon?: React.ReactNode;
  tone?: 'brand' | 'danger';
};
type ConfirmReq = ConfirmOptions & { id: number; resolve: (v: boolean) => void };
const confirmListeners = new Set<(r: ConfirmReq) => void>();
let confirmSeq = 0;

/** Dialog konfirmasi modal. Tunggu jawabannya: if (!(await confirmAction({ title, tone: 'danger' }))) return; */
export const confirmAction = (o: ConfirmOptions) =>
  new Promise<boolean>((resolve) => {
    const req: ConfirmReq = { ...o, id: ++confirmSeq, resolve };
    confirmListeners.forEach((l) => l(req));
  });

const ConfirmHost: React.FC = () => {
  const [queue, setQueue] = useState<ConfirmReq[]>([]);
  useEffect(() => {
    const on = (r: ConfirmReq) => setQueue((p) => [...p, r]);
    confirmListeners.add(on);
    return () => { confirmListeners.delete(on); };
  }, []);
  const head = queue[0];
  if (!head) return null;
  const settle = (v: boolean) => { head.resolve(v); setQueue((p) => p.slice(1)); };
  const danger = head.tone === 'danger';
  return (
    <Modal
      title={head.title}
      error={null}
      onClose={() => settle(false)}
      onSubmit={(e) => { e.preventDefault(); settle(true); }}
      submitLabel={head.confirmLabel ?? 'Lanjutkan'}
      tone={danger ? 'danger' : 'brand'}
      icon={head.icon ?? (danger ? <IconAlert className="h-5 w-5" /> : <IconInfo className="h-5 w-5" />)}
    >
      {head.body && <p className="confirm-body">{head.body}</p>}
    </Modal>
  );
};

export const Toaster: React.FC = () => (
  <>
    <ToastViewport />
    <ConfirmHost />
  </>
);

/* ---------- form field ---------- */
export const Field: React.FC<{ label: string; htmlFor?: string; hint?: string; error?: string; children: React.ReactNode }> = ({ label, htmlFor, hint, error, children }) => (
  <div>
    <label className="label" htmlFor={htmlFor}>{label}</label>
    {children}
    {error ? <p className="field-error" role="alert">{error}</p> : hint && <p className="field-hint">{hint}</p>}
  </div>
);

/* ---------- select kustom ---------- */
// Pengganti <select> native: popover di desktop, bottom-sheet di layar sempit.
// Fokus tetap di tombol (pola combobox), jadi focus-trap Modal tetap bekerja.
export interface SelectOption { value: string; label: string; disabled?: boolean }

export const Select: React.FC<{
  id?: string;
  value: string;
  onChange: (value: string) => void;
  options: SelectOption[];
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  'aria-label'?: string;
}> = ({ id, value, onChange, options, placeholder = 'Pilih…', required, disabled, ...rest }) => {
  const btn = useRef<HTMLButtonElement>(null);
  const list = useRef<HTMLUListElement>(null);
  const [open, setOpen] = useState(false);
  const [active, setActive] = useState(-1);
  const [pos, setPos] = useState<{ left: number; width: number; top?: number; bottom?: number; maxH: number } | null>(null);
  const [sheet, setSheet] = useState(false);
  const uid = React.useId();
  const listId = `${uid}-list`;

  const selected = options.find((o) => o.value === value);
  const enabledIdx = (from: number, dir: 1 | -1) => {
    for (let i = from; i >= 0 && i < options.length; i += dir) if (!options[i].disabled) return i;
    return -1;
  };

  const place = () => {
    const el = btn.current;
    if (!el) return;
    const narrow = window.matchMedia('(max-width: 639px)').matches;
    setSheet(narrow);
    if (narrow) return;
    const r = el.getBoundingClientRect();
    const gap = 6, margin = 12;
    const below = window.innerHeight - r.bottom - margin;
    const above = r.top - margin;
    const flip = below < 200 && above > below;
    setPos({
      left: r.left, width: r.width,
      ...(flip ? { bottom: window.innerHeight - r.top + gap } : { top: r.bottom + gap }),
      maxH: Math.max(160, Math.min(320, (flip ? above : below) - gap)),
    });
  };

  const openList = () => {
    if (disabled) return;
    place();
    setActive(Math.max(0, options.findIndex((o) => o.value === value)));
    setOpen(true);
  };
  const close = () => setOpen(false);
  const choose = (i: number) => {
    const o = options[i];
    if (!o || o.disabled) return;
    onChange(o.value);
    close();
    btn.current?.focus();
  };

  // tutup saat klik di luar; ikuti posisi saat scroll / resize
  useEffect(() => {
    if (!open) return;
    const onDown = (e: PointerEvent) => {
      const t = e.target as Node;
      if (btn.current?.contains(t) || list.current?.closest('.select-layer')?.contains(t)) return;
      close();
    };
    const onMove = () => place();
    document.addEventListener('pointerdown', onDown);
    window.addEventListener('resize', onMove);
    window.addEventListener('scroll', onMove, true);
    return () => {
      document.removeEventListener('pointerdown', onDown);
      window.removeEventListener('resize', onMove);
      window.removeEventListener('scroll', onMove, true);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  // sheet mobile mengunci scroll halaman
  useEffect(() => {
    if (!open || !sheet) return;
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = prev; };
  }, [open, sheet]);

  useEffect(() => {
    if (open) list.current?.querySelector<HTMLElement>('[data-active="true"]')?.scrollIntoView({ block: 'nearest' });
  }, [open, active]);

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (disabled) return;
    if (!open) {
      if (['ArrowDown', 'ArrowUp', 'Enter', ' '].includes(e.key)) { e.preventDefault(); openList(); }
      return;
    }
    switch (e.key) {
      case 'Escape': e.preventDefault(); e.stopPropagation(); close(); break;
      case 'Tab': close(); break;
      case 'ArrowDown': e.preventDefault(); { const n = enabledIdx(active + 1, 1); if (n >= 0) setActive(n); } break;
      case 'ArrowUp': e.preventDefault(); { const n = enabledIdx(active - 1, -1); if (n >= 0) setActive(n); } break;
      case 'Home': e.preventDefault(); setActive(Math.max(0, enabledIdx(0, 1))); break;
      case 'End': e.preventDefault(); { const n = enabledIdx(options.length - 1, -1); if (n >= 0) setActive(n); } break;
      case 'Enter': case ' ': e.preventDefault(); choose(active); break;
      default:
        if (e.key.length === 1) {
          const k = e.key.toLowerCase();
          const n = options.findIndex((o, i) => i > active && !o.disabled && o.label.toLowerCase().startsWith(k));
          const m = n >= 0 ? n : options.findIndex((o) => !o.disabled && o.label.toLowerCase().startsWith(k));
          if (m >= 0) setActive(m);
        }
    }
  };

  const items = (
    <ul ref={list} id={listId} role="listbox" aria-label={rest['aria-label'] ?? placeholder} className="select-list"
      style={!sheet && pos ? { maxHeight: pos.maxH } : undefined}>
      {options.length === 0 && <li className="select-empty">Tidak ada pilihan</li>}
      {options.map((o, i) => (
        <li
          key={o.value}
          id={`${uid}-o${i}`}
          role="option"
          aria-selected={o.value === value}
          aria-disabled={o.disabled || undefined}
          data-active={i === active}
          className="select-opt"
          onMouseDown={(e) => e.preventDefault()}
          onMouseMove={() => !o.disabled && setActive(i)}
          onClick={() => choose(i)}
        >
          <span className="truncate">{o.label}</span>
          {o.value === value && <IconCheckPlain className="h-4 w-4 shrink-0" />}
        </li>
      ))}
    </ul>
  );

  return (
    <div className="select-wrap">
      <button
        ref={btn}
        id={id}
        type="button"
        className="input select-trigger"
        data-open={open}
        data-empty={!selected}
        disabled={disabled}
        role="combobox"
        aria-haspopup="listbox"
        aria-expanded={open}
        aria-controls={open ? listId : undefined}
        aria-activedescendant={open && active >= 0 ? `${uid}-o${active}` : undefined}
        aria-label={rest['aria-label']}
        onClick={() => (open ? close() : openList())}
        onKeyDown={onKeyDown}
        onBlur={() => { if (!sheet) close(); }}
      >
        <span className="truncate">{selected ? selected.label : placeholder}</span>
        <IconChevronDown className="select-chevron h-4 w-4 shrink-0" />
      </button>
      {/* penampung validasi native: memunculkan "wajib diisi" tepat di bawah tombol */}
      {required && <input className="select-proxy" tabIndex={-1} aria-hidden="true" required value={value} onChange={() => {}} />}
      {open && createPortal(
        <div className={`select-layer ${sheet ? 'is-sheet' : ''}`}>
          {sheet && <div className="select-backdrop" onClick={close} />}
          <div className="select-pop" onMouseDown={(e) => e.preventDefault()} style={!sheet && pos ? { left: pos.left, width: pos.width, top: pos.top, bottom: pos.bottom } : undefined}>
            {sheet && (
              <div className="select-sheet-head">
                <span className="select-grip" />
                <p>{rest['aria-label'] ?? placeholder}</p>
              </div>
            )}
            {items}
          </div>
        </div>,
        document.body
      )}
    </div>
  );
};

/* ---------- baris alat di dalam kartu tabel ---------- */
// Judul, jumlah data, pencarian/filter, dan tombol aksi utama duduk di dalam kartu tabel.
export const TableBar: React.FC<{
  title: string; count?: number; noun?: string; action?: React.ReactNode; children?: React.ReactNode;
}> = ({ title, count, noun, action, children }) => (
  <div className="table-bar">
    <div className="table-bar-main">
      <div className="table-bar-head">
        <h2 className="table-title">{title}</h2>
        {count !== undefined && (
          <span className="table-count"><span className="num">{count}</span> {noun}</span>
        )}
      </div>
      {children}
    </div>
    {action && <div className="table-bar-action">{action}</div>}
  </div>
);

export const TableFoot: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <div className="table-foot">{children}</div>
);

/* ---------- filter status: tab pil di layar lebar, select di layar sempit ---------- */
export const FilterTabs: React.FC<{
  label: string; value: string; onChange: (v: string) => void; options: SelectOption[];
}> = ({ label, value, onChange, options }) => (
  <>
    <div className="chips hidden md:flex" role="group" aria-label={label}>
      {options.map((o) => (
        <button key={o.value || 'all'} type="button" className="chip" aria-pressed={value === o.value} onClick={() => onChange(o.value)}>
          {o.label}
        </button>
      ))}
    </div>
    <div className="filter-select md:hidden">
      <Select aria-label={label} value={value} onChange={onChange} options={options} />
    </div>
  </>
);

/* ---------- modal (dialog di desktop, bottom-sheet di mobile) ---------- */
const FOCUSABLE = 'a[href],button:not(:disabled),input:not(:disabled):not(.select-proxy),select:not(:disabled),textarea:not(:disabled),[tabindex]:not([tabindex="-1"])';

export const Modal: React.FC<{
  title: string;
  subtitle?: string;
  icon?: React.ReactNode;
  tone?: 'brand' | 'danger';
  onClose: () => void;
  onSubmit: (e: React.FormEvent) => void;
  submitLabel: string;
  error?: string | null;
  busy?: boolean;
  children: React.ReactNode;
}> = ({ title, subtitle, icon, tone = 'brand', onClose, onSubmit, submitLabel, error, busy, children }) => {
  const panel = useRef<HTMLFormElement>(null);
  const onCloseRef = useRef(onClose);
  onCloseRef.current = onClose;

  useEffect(() => {
    const opener = document.activeElement as HTMLElement | null;
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    // fokus ke field pertama; kalau tidak ada, ke panel
    const first = panel.current?.querySelector<HTMLElement>('.modal-body input:not(.select-proxy), .modal-body .select-trigger, .modal-body select, .modal-body textarea');
    (first ?? panel.current)?.focus();

    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { onCloseRef.current(); return; }
      if (e.key !== 'Tab' || !panel.current) return;
      const nodes = Array.from(panel.current.querySelectorAll<HTMLElement>(FOCUSABLE));
      if (!nodes.length) return;
      const a = nodes[0], z = nodes[nodes.length - 1];
      if (!panel.current.contains(document.activeElement)) { e.preventDefault(); a.focus(); return; }
      if (e.shiftKey && document.activeElement === a) { e.preventDefault(); z.focus(); }
      else if (!e.shiftKey && document.activeElement === z) { e.preventDefault(); a.focus(); }
    };
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = prevOverflow;
      opener?.focus?.();
    };
  }, []);

  // Portal ke body: modal tidak ikut terkena margin `space-y-*` dari container halaman.
  return createPortal(
    <div className="modal-root" role="dialog" aria-modal="true" aria-label={title}>
      <div className="modal-backdrop" onClick={onClose} />
      <form ref={panel} tabIndex={-1} className="modal-panel outline-none" onSubmit={(e) => { if (busy) { e.preventDefault(); return; } onSubmit(e); }}>
        <div className="modal-head">
          <div className="modal-head-main">
            {icon && <span className="modal-icon" data-tone={tone}>{icon}</span>}
            <div className="modal-head-text">
              <h2>{title}</h2>
              {subtitle && <p className="modal-sub">{subtitle}</p>}
            </div>
          </div>
          <button type="button" className="icon-btn" onClick={onClose} aria-label="Tutup">
            <IconX className="h-5 w-5" />
          </button>
        </div>
        <div className="modal-body space-y-4">
          {error && <ErrorAlert message={error} />}
          {children}
        </div>
        <div className="modal-foot">
          <button type="button" className="btn btn-ghost" onClick={onClose}>Batal</button>
          <button type="submit" className={tone === 'danger' ? 'btn btn-danger-solid' : 'btn btn-primary'} aria-disabled={busy || undefined}>
            {busy && <span className="spinner" aria-hidden="true" />}
            {submitLabel}
          </button>
        </div>
      </form>
    </div>,
    document.body
  );
};

/* ---------- pagination ---------- */
export const Pagination: React.FC<{
  page: number; totalPages: number; total: number; noun: string; onPage: (p: number) => void;
}> = ({ page, totalPages, total, noun, onPage }) => {
  const last = totalPages || 1;
  return (
    <div className="pager flex items-center justify-between gap-3">
      <p className="meta">
        <span className="num font-medium" style={{ color: 'var(--ink)' }}>{total}</span> {noun}
      </p>
      <div className="flex items-center gap-2">
        <span className="meta num hidden sm:inline">Halaman {page} dari {last}</span>
        <button className="btn btn-ghost btn-sm" disabled={page <= 1} onClick={() => onPage(page - 1)} aria-label="Halaman sebelumnya">
          <IconChevronLeft className="h-4 w-4" /> <span className="hidden sm:inline">Sebelumnya</span>
        </button>
        <span className="meta num sm:hidden">{page}/{last}</span>
        <button className="btn btn-ghost btn-sm" disabled={page >= last} onClick={() => onPage(page + 1)} aria-label="Halaman berikutnya">
          <span className="hidden sm:inline">Berikutnya</span> <IconChevronRight className="h-4 w-4" />
        </button>
      </div>
    </div>
  );
};

export const initials = (name?: string) =>
  (name ?? '?').split(/\s+/).filter(Boolean).slice(0, 2).map((s) => s[0]?.toUpperCase()).join('') || '?';

/* ---------- command palette (modern search: click / Cmd+K) ---------- */
export interface CommandItem {
  label: string;
  description?: string;
  shortcut?: string;
  icon?: React.ReactNode;
  onSelect: () => void;
  section?: string;
  disabled?: boolean;
}

const highlight = (text: string, q: string) => {
  const i = q ? text.toLowerCase().indexOf(q.toLowerCase()) : -1;
  if (i < 0) return text;
  return (<>{text.slice(0, i)}<mark className="cmd-mark">{text.slice(i, i + q.length)}</mark>{text.slice(i + q.length)}</>);
};

export const CommandPalette: React.FC<{
  isOpen: boolean;
  onClose: () => void;
  items: CommandItem[];
  placeholder?: string;
  title?: string;
}> = ({ isOpen, onClose, items, placeholder = 'Cari halaman atau perintah', title }) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLUListElement>(null);
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  useEffect(() => {
    if (!isOpen) return;
    setQuery('');
    setSelectedIndex(0);
    inputRef.current?.focus();
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => { document.body.style.overflow = prev; };
  }, [isOpen]);

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return items.filter((i) => !i.disabled && (!q || i.label.toLowerCase().includes(q) || (i.description ?? '').toLowerCase().includes(q)));
  }, [items, query]);

  // kelompokkan per bagian, urutan asli dipertahankan; indeks tetap berlaku untuk daftar datar
  const groups = useMemo(() => {
    const out: { name: string; rows: { item: CommandItem; idx: number }[] }[] = [];
    filtered.forEach((item, idx) => {
      const name = item.section ?? '';
      let g = out.find((x) => x.name === name);
      if (!g) { g = { name, rows: [] }; out.push(g); }
      g.rows.push({ item, idx });
    });
    return out;
  }, [filtered]);

  // item yang dipilih lewat keyboard selalu terlihat
  useEffect(() => {
    listRef.current?.querySelector('[aria-selected="true"]')?.scrollIntoView({ block: 'nearest' });
  }, [selectedIndex, filtered]);

  const run = (item?: CommandItem) => { if (!item) return; onClose(); item.onSelect(); };

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Escape') { e.preventDefault(); onClose(); return; }
    if (e.key === 'ArrowDown') { e.preventDefault(); setSelectedIndex((i) => (filtered.length ? (i + 1) % filtered.length : 0)); }
    else if (e.key === 'ArrowUp') { e.preventDefault(); setSelectedIndex((i) => (filtered.length ? (i - 1 + filtered.length) % filtered.length : 0)); }
    else if (e.key === 'Home') { e.preventDefault(); setSelectedIndex(0); }
    else if (e.key === 'End') { e.preventDefault(); setSelectedIndex(Math.max(0, filtered.length - 1)); }
    else if (e.key === 'Enter') { e.preventDefault(); run(filtered[selectedIndex]); }
  };

  if (!isOpen) return null;
  const q = query.trim();

  return createPortal(
    <div className="cmd-root" role="dialog" aria-modal="true" aria-label={title ?? 'Pencarian cepat'}>
      <div className="cmd-backdrop" onClick={onClose} />
      <div className="cmd-panel" onKeyDown={onKeyDown}>
        <div className="cmd-input-wrap">
          <IconSearch className="cmd-icon" aria-hidden="true" />
          <input
            ref={inputRef}
            type="text"
            role="combobox"
            aria-expanded="true"
            aria-controls="cmd-list"
            aria-activedescendant={filtered[selectedIndex] ? `cmd-opt-${selectedIndex}` : undefined}
            value={query}
            onChange={(e) => { setQuery(e.target.value); setSelectedIndex(0); }}
            placeholder={placeholder}
            autoComplete="off"
            spellCheck={false}
            aria-label={placeholder}
          />
          <button type="button" className="cmd-esc" onClick={onClose} aria-label="Tutup">Esc</button>
        </div>

        <ul className="cmd-list" id="cmd-list" role="listbox" ref={listRef}>
          {groups.map((g) => (
            <li key={g.name || 'x'} role="presentation" className="cmd-group">
              {g.name && <p className="cmd-group-name" role="presentation">{g.name}</p>}
              <ul role="presentation">
                {g.rows.map(({ item, idx }) => (
                  <li
                    key={item.label}
                    id={`cmd-opt-${idx}`}
                    role="option"
                    aria-selected={idx === selectedIndex}
                    className={`cmd-item ${idx === selectedIndex ? 'selected' : ''}`}
                    onClick={() => run(item)}
                    onMouseMove={() => idx !== selectedIndex && setSelectedIndex(idx)}
                  >
                    {item.icon && <span className="cmd-item-icon" aria-hidden="true">{item.icon}</span>}
                    <span className="cmd-item-text">
                      <span className="cmd-item-label">{highlight(item.label, q)}</span>
                      {item.description && <span className="cmd-item-desc">{item.description}</span>}
                    </span>
                    {item.shortcut && (
                      <span className="cmd-keys" aria-hidden="true">
                        {item.shortcut.split(' ').map((k, n) => <kbd key={n}>{k}</kbd>)}
                      </span>
                    )}
                  </li>
                ))}
              </ul>
            </li>
          ))}
          {filtered.length === 0 && (
            <li className="cmd-empty" role="presentation">
              <p>Tidak ada hasil untuk “{q}”</p>
              <p>Coba kata lain, misalnya nama halaman seperti “aset” atau “tiket”.</p>
            </li>
          )}
        </ul>

        <div className="cmd-foot" aria-hidden="true">
          <span><kbd>↑</kbd><kbd>↓</kbd> pilih</span>
          <span><kbd>Enter</kbd> buka</span>
          <span><kbd>Esc</kbd> tutup</span>
        </div>
      </div>
    </div>,
    document.body
  );
};

/* ---------- notification dropdown ---------- */
export interface NotificationItem {
  id: string;
  title: string;
  message?: string;
  time: string;
  unread?: boolean;
  onClick?: () => void;
}

export const NotificationDropdown: React.FC<{
  items: NotificationItem[];
  onMarkAllRead?: () => void;
  onViewAll?: () => void;
}> = ({ items, onMarkAllRead, onViewAll }) => {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const btnRef = useRef<HTMLButtonElement>(null);
  const unread = items.filter((i) => i.unread).length;

  useEffect(() => {
    if (!open) return;
    const onDown = (e: PointerEvent) => {
      if (!rootRef.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') { setOpen(false); btnRef.current?.focus(); }
    };
    document.addEventListener('pointerdown', onDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('pointerdown', onDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  return (
    <div className="notif" ref={rootRef}>
      <button
        ref={btnRef}
        type="button"
        className="tool-btn notif-btn"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-label={unread > 0 ? `Notifikasi, ${unread} belum dibaca` : 'Notifikasi'}
        title="Notifikasi"
        onClick={() => setOpen((v) => !v)}
      >
        <IconBell className="h-[18px] w-[18px]" />
        {unread > 0 && <span className="notif-badge num" aria-hidden="true">{unread}</span>}
      </button>

      {open && (
        <div className="notif-panel" role="dialog" aria-label="Notifikasi">
          <div className="notif-head">
            <h3>Notifikasi</h3>
            {unread > 0 && onMarkAllRead && (
              <button type="button" className="notif-mark-all" onClick={onMarkAllRead}>Tandai semua dibaca</button>
            )}
          </div>
          <ul className="notif-list">
            {items.length === 0 ? (
              <li className="notif-empty">Belum ada notifikasi</li>
            ) : (
              items.map((item) => (
                <li key={item.id}>
                  <button
                    type="button"
                    className={`notif-item ${item.unread ? 'unread' : ''}`}
                    onClick={() => { item.onClick?.(); setOpen(false); }}
                  >
                    <span className="notif-item-content">
                      <span className="notif-item-title">{item.title}</span>
                      {item.message && <span className="notif-item-msg">{item.message}</span>}
                      <time className="notif-item-time">{item.time}</time>
                    </span>
                    {item.unread && <span className="notif-dot" aria-label="Belum dibaca" />}
                  </button>
                </li>
              ))
            )}
          </ul>
          {onViewAll && (
            <button type="button" className="notif-view-all" onClick={() => { onViewAll(); setOpen(false); }}>Lihat semua</button>
          )}
        </div>
      )}
    </div>
  );
};

/* ---------- theme toggle button (standalone) ---------- */
export const ThemeToggle: React.FC<{
  dark: boolean;
  onToggle: () => void;
  'aria-label'?: string;
}> = ({ dark, onToggle, 'aria-label': ariaLabel }) => (
  <button onClick={onToggle} className="tool-btn" aria-label={ariaLabel ?? (dark ? 'Mode terang' : 'Mode gelap')} title={ariaLabel ?? (dark ? 'Mode terang' : 'Mode gelap')}>
    {dark ? <IconSun className="h-5 w-5" /> : <IconMoon className="h-5 w-5" />}
  </button>
);
