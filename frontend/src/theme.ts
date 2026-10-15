// Tema: ikuti pilihan tersimpan, kalau belum ada ikuti preferensi sistem.
// Dipanggil sebelum render pertama (main.tsx) supaya tidak ada kilatan tema salah.
const KEY = 'tp-theme';
export type Theme = 'light' | 'dark';

export function storedTheme(): Theme | null {
  try {
    const v = localStorage.getItem(KEY);
    return v === 'dark' || v === 'light' ? v : null;
  } catch {
    return null;
  }
}

export function systemTheme(): Theme {
  return window.matchMedia?.('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export function applyTheme(t: Theme) {
  const root = document.documentElement;
  if (t === 'dark') root.dataset.theme = 'dark';
  else delete root.dataset.theme;
}

export function saveTheme(t: Theme) {
  try { localStorage.setItem(KEY, t); } catch { /* abaikan */ }
}

export function initTheme() {
  applyTheme(storedTheme() ?? systemTheme());
}
