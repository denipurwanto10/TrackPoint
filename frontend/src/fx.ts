// Cahaya lembut yang mengikuti kursor di atas kartu (hanya perangkat dengan hover).
// Satu listener terdelegasi, jadi kartu baru otomatis ikut tanpa kode tambahan.
export function initPointerGlow() {
  if (!window.matchMedia?.('(hover: hover)').matches) return;
  if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;
  let raf = 0;
  document.addEventListener('pointermove', (e) => {
    cancelAnimationFrame(raf);
    raf = requestAnimationFrame(() => {
      const el = (e.target as HTMLElement | null)?.closest?.('.card, .stat, .table-card, .promo') as HTMLElement | null;
      if (!el) return;
      const r = el.getBoundingClientRect();
      el.style.setProperty('--mx', `${e.clientX - r.left}px`);
      el.style.setProperty('--my', `${e.clientY - r.top}px`);
    });
  }, { passive: true });
}
