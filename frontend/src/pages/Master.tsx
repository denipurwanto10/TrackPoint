import React, { useEffect, useState } from 'react';
import { apiFetch, type ApiError } from '../api';
import { IconPlus, IconTrash } from '../components/Icons';
import { Empty, ErrorAlert, PageHeader, TableSkeleton, confirmAction, toast } from '../components/ui';

interface Cat { id: string; nama: string; deskripsi?: string | null; is_active: boolean }
interface Loc { id: string; nama: string; gedung?: string | null; is_active: boolean }

const Master: React.FC = () => {
  const [cats, setCats] = useState<Cat[]>([]);
  const [locs, setLocs] = useState<Loc[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [catForm, setCatForm] = useState({ nama: '', deskripsi: '' });
  const [locForm, setLocForm] = useState({ nama: '', gedung: '' });

  const load = async () => {
    try {
      const [c, l] = await Promise.all([
        apiFetch('/api/v1/categories'),
        apiFetch('/api/v1/locations'),
      ]);
      setCats(c);
      setLocs(l);
      setErr(null);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(); }, []);

  const addCat = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await apiFetch('/api/v1/categories', { method: 'POST', body: JSON.stringify(catForm) });
      setCatForm({ nama: '', deskripsi: '' });
      toast('Kategori ditambahkan');
      load();
    } catch (e) { setErr((e as ApiError).message); }
  };
  const addLoc = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await apiFetch('/api/v1/locations', { method: 'POST', body: JSON.stringify(locForm) });
      setLocForm({ nama: '', gedung: '' });
      toast('Lokasi ditambahkan');
      load();
    } catch (e) { setErr((e as ApiError).message); }
  };
  const delCat = async (id: string) => {
    if (!(await confirmAction({
      title: 'Hapus kategori?',
      body: 'Kategori dihapus dari data master sehingga tidak bisa dipakai lagi untuk aset baru.',
      confirmLabel: 'Hapus kategori',
      icon: <IconTrash className="h-5 w-5" />,
      tone: 'danger',
    }))) return;
    try { await apiFetch(`/api/v1/categories/${id}`, { method: 'DELETE' }); toast('Kategori dihapus'); load(); }
    catch (e) { setErr((e as ApiError).message); }
  };
  const delLoc = async (id: string) => {
    if (!(await confirmAction({
      title: 'Hapus lokasi?',
      body: 'Lokasi dihapus dari data master sehingga tidak bisa dipilih lagi saat menambah aset.',
      confirmLabel: 'Hapus lokasi',
      icon: <IconTrash className="h-5 w-5" />,
      tone: 'danger',
    }))) return;
    try { await apiFetch(`/api/v1/locations/${id}`, { method: 'DELETE' }); toast('Lokasi dihapus'); load(); }
    catch (e) { setErr((e as ApiError).message); }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Kategori & lokasi" subtitle="Data master untuk mengelompokkan dan menempatkan aset." />
      {err && <ErrorAlert message={err} />}

      <div className="grid grid-cols-1 items-start gap-5 lg:grid-cols-2">
        <section aria-labelledby="kategori" className="card">
          <div className="mb-4 flex items-baseline justify-between gap-3">
            <h2 id="kategori" className="section-title">Kategori</h2>
            <span className="meta num">{cats.length} kategori</span>
          </div>
          <form onSubmit={addCat} className="flex gap-2">
            <input required value={catForm.nama} onChange={(e) => setCatForm({ ...catForm, nama: e.target.value })} placeholder="Nama kategori baru" className="input" aria-label="Nama kategori" />
            <button type="submit" className="btn btn-primary shrink-0" aria-label="Tambah kategori"><IconPlus className="h-4 w-4" /><span className="hidden sm:inline">Tambah</span></button>
          </form>
          <div className="mt-4">
            {loading ? <TableSkeleton rows={4} /> : cats.length === 0 ? (
              <Empty text="Belum ada kategori" hint="Tambahkan kategori seperti Laptop atau Printer." />
            ) : (
              <ul className="border-t" style={{ borderColor: 'var(--line)' }}>
                {cats.map((c) => (
                  <li key={c.id} className="flex items-center justify-between gap-3 border-b py-1.5 pl-0.5" style={{ borderColor: 'var(--line)' }}>
                    <span className="flex min-w-0 items-center gap-3 text-sm font-medium">
                      <span className="truncate">{c.nama}</span>
                      {!c.is_active && <span className="badge badge-mute shrink-0">Nonaktif</span>}
                    </span>
                    <button onClick={() => delCat(c.id)} className="icon-btn danger" aria-label={`Hapus ${c.nama}`}><IconTrash className="h-[18px] w-[18px]" /></button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>

        <section aria-labelledby="lokasi" className="card">
          <div className="mb-4 flex items-baseline justify-between gap-3">
            <h2 id="lokasi" className="section-title">Lokasi</h2>
            <span className="meta num">{locs.length} lokasi</span>
          </div>
          <form onSubmit={addLoc} className="flex flex-col gap-2 sm:flex-row">
            <input required value={locForm.nama} onChange={(e) => setLocForm({ ...locForm, nama: e.target.value })} placeholder="Nama lokasi baru" className="input" aria-label="Nama lokasi" />
            <input value={locForm.gedung} onChange={(e) => setLocForm({ ...locForm, gedung: e.target.value })} placeholder="Gedung" className="input sm:max-w-[9rem]" aria-label="Gedung" />
            <button type="submit" className="btn btn-primary shrink-0"><IconPlus className="h-4 w-4" /> Tambah</button>
          </form>
          <div className="mt-4">
            {loading ? <TableSkeleton rows={4} /> : locs.length === 0 ? (
              <Empty text="Belum ada lokasi" hint="Tambahkan ruangan atau lantai tempat aset disimpan." />
            ) : (
              <ul className="border-t" style={{ borderColor: 'var(--line)' }}>
                {locs.map((l) => (
                  <li key={l.id} className="flex items-center justify-between gap-3 border-b py-1.5 pl-0.5" style={{ borderColor: 'var(--line)' }}>
                    <span className="flex min-w-0 items-baseline gap-3 text-sm font-medium">
                      <span className="truncate">{l.nama}</span>
                      {l.gedung && <span className="meta shrink-0 font-normal">{l.gedung}</span>}
                      {!l.is_active && <span className="badge badge-mute shrink-0">Nonaktif</span>}
                    </span>
                    <button onClick={() => delLoc(l.id)} className="icon-btn danger" aria-label={`Hapus ${l.nama}`}><IconTrash className="h-[18px] w-[18px]" /></button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>
      </div>
    </div>
  );
};

export default Master;
