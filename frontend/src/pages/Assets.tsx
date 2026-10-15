import React, { useEffect, useState } from 'react';
import { apiFetch, apiPage, type ApiError } from '../api';
import { useAuth } from '../auth';
import { IconEdit, IconPlus, IconSearch, IconSwap } from '../components/Icons';
import { Empty, ErrorAlert, Field, Modal, PageHeader, FilterTabs, Pagination, Select, StatusBadge, TableBar, TableFoot, TableSkeleton, statusLabel, toast } from '../components/ui';

interface Asset {
  id: string; kode: string; nama: string;
  serial_number?: string | null;
  category_id: string; location_id: string;
  holder_id?: string | null;
  status: string;
  purchase_date?: string | null;
  catatan?: string | null;
  category_nama: string; location_nama: string;
  holder_nama?: string | null;
}
interface Opt { id: string; nama: string }

const STATUSES = ['', 'AVAILABLE', 'ASSIGNED', 'IN_MAINTENANCE', 'INACTIVE'];

const Assets: React.FC = () => {
  const { user } = useAuth();
  const canWrite = user?.role === 'ADMIN' || user?.role === 'TEKNISI';
  const [items, setItems] = useState<Asset[]>([]);
  const [meta, setMeta] = useState<{ page: number; total_pages: number; total: number }>({ page: 1, total_pages: 0, total: 0 });
  const [q, setQ] = useState('');
  const [status, setStatus] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [cats, setCats] = useState<Opt[]>([]);
  const [locs, setLocs] = useState<Opt[]>([]);
  const [showForm, setShowForm] = useState(false);
  const [editing, setEditing] = useState<Asset | null>(null);
  const [form, setForm] = useState({ nama: '', serial_number: '', category_id: '', location_id: '', purchase_date: '', catatan: '' });
  const [moveId, setMoveId] = useState<string | null>(null);
  const [move, setMove] = useState({ to_location_id: '', to_holder_id: '', catatan: '' });

  const load = async (page = 1, s = status) => {
    try {
      const params = new URLSearchParams({ page: String(page), per_page: '20' });
      if (q) params.set('q', q);
      if (s) params.set('status', s);
      const { items, meta } = await apiPage(`/api/v1/assets?${params}`);
      setItems(items);
      setMeta({ page: meta.page, total_pages: meta.total_pages, total: meta.total });
      setErr(null);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(1);
    apiFetch('/api/v1/categories?active=true').then(setCats).catch(() => undefined);
    apiFetch('/api/v1/locations?active=true').then(setLocs).catch(() => undefined);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const openCreate = () => {
    setErr(null);
    setEditing(null);
    setForm({ nama: '', serial_number: '', category_id: '', location_id: '', purchase_date: '', catatan: '' });
    setShowForm(true);
  };
  const openEdit = (a: Asset) => {
    setErr(null);
    setEditing(a);
    setForm({
      nama: a.nama, serial_number: a.serial_number ?? '',
      category_id: a.category_id, location_id: a.location_id,
      purchase_date: a.purchase_date ? a.purchase_date.slice(0, 10) : '',
      catatan: a.catatan ?? '',
    });
    setShowForm(true);
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    // Custom validation for required Select fields (hidden proxy doesn't show browser UI)
    if (!form.nama.trim()) { setErr('Nama aset wajib diisi'); setSaving(false); return; }
    if (!form.category_id) { setErr('Kategori wajib dipilih'); setSaving(false); return; }
    if (!editing && !form.location_id) { setErr('Lokasi awal wajib dipilih'); setSaving(false); return; }
    setSaving(true);
    setErr(null);
    try {
      const payload = {
        nama: form.nama,
        serial_number: form.serial_number || undefined,
        category_id: form.category_id || undefined,
        location_id: editing ? undefined : form.location_id || undefined,
        purchase_date: form.purchase_date || undefined,
        catatan: form.catatan || undefined,
      };
      if (editing) await apiFetch(`/api/v1/assets/${editing.id}`, { method: 'PATCH', body: JSON.stringify(payload) });
      else await apiFetch('/api/v1/assets', { method: 'POST', body: JSON.stringify(payload) });
      setShowForm(false);
      toast(editing ? 'Perubahan disimpan' : 'Aset ditambahkan');
      load(meta.page);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setSaving(false);
    }
  };

  const submitMove = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!moveId) return;
    // Custom validation for required Select fields
    if (!move.to_location_id) { setErr('Lokasi tujuan wajib dipilih'); setSaving(false); return; }
    setSaving(true);
    setErr(null);
    try {
      await apiFetch(`/api/v1/assets/${moveId}/movements`, {
        method: 'POST',
        body: JSON.stringify({
          to_location_id: move.to_location_id,
          to_holder_id: move.to_holder_id || undefined,
          catatan: move.catatan || undefined,
        }),
      });
      setMoveId(null);
      setMove({ to_location_id: '', to_holder_id: '', catatan: '' });
      toast('Mutasi disimpan');
      load(meta.page);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setSaving(false);
    }
  };

  const modalOpen = showForm || !!moveId;
  const filtered = !!q || !!status;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Aset"
        subtitle="Seluruh perangkat dan inventaris IT beserta lokasi dan pemegangnya."
      />
      {err && !modalOpen && <ErrorAlert message={err} />}

      <div className="table-card">
        <TableBar
          title="Daftar aset"
          count={loading ? undefined : meta.total}
          noun="aset"
          action={canWrite && (
            <button onClick={openCreate} className="btn btn-primary">
              <IconPlus className="h-4 w-4" /> Tambah aset
            </button>
          )}
        >
          <form onSubmit={(e) => { e.preventDefault(); load(1); }} className="table-search">
            <div className="search-wrap">
              <IconSearch className="h-[18px] w-[18px]" />
              <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="Cari kode atau nama" className="input" aria-label="Cari aset" />
            </div>
            <button type="submit" className="btn btn-ghost">Cari</button>
          </form>
          <FilterTabs label="Filter status" value={status} onChange={(s) => { setStatus(s); load(1, s); }}
            options={STATUSES.map((s) => ({ value: s, label: s ? statusLabel(s) : 'Semua status' }))} />
        </TableBar>
        {loading ? <TableSkeleton /> : items.length === 0 ? (
          <Empty
            text={filtered ? 'Tidak ada aset yang cocok' : 'Belum ada aset'}
            hint={filtered ? 'Ubah kata kunci atau filter status.' : canWrite ? 'Tambahkan aset pertama untuk mulai melacak.' : undefined}
            action={!filtered && canWrite ? <button onClick={openCreate} className="btn btn-primary"><IconPlus className="h-4 w-4" /> Tambah aset</button> : undefined}
          />
        ) : (
          <table className="rtable">
            <thead>
              <tr>
                <th>Aset</th>
                <th>Kategori</th>
                <th>Lokasi</th>
                <th>Pemegang</th>
                <th>Status</th>
                {canWrite && <th><span className="sr-only">Aksi</span></th>}
              </tr>
            </thead>
            <tbody>
              {items.map((a) => (
                <tr key={a.id}>
                  <td className="cell-main" data-label="Aset">
                    <div className="font-medium">{a.nama}</div>
                    <span className="code mt-1 inline-block">{a.kode}</span>
                  </td>
                  <td data-label="Kategori">{a.category_nama}</td>
                  <td data-label="Lokasi">{a.location_nama}</td>
                  <td data-label="Pemegang" style={a.holder_nama ? undefined : { color: 'var(--muted)' }}>{a.holder_nama ?? '—'}</td>
                  <td data-label="Status"><StatusBadge value={a.status} /></td>
                  {canWrite && (
                    <td className="cell-actions" data-label="Aksi">
                      <div className="row-actions">
                        <button onClick={() => openEdit(a)} className="btn btn-ghost btn-sm btn-row"><IconEdit className="h-4 w-4" /> Ubah</button>
                        <button
                          onClick={() => { setErr(null); setMoveId(a.id); setMove({ to_location_id: a.location_id, to_holder_id: '', catatan: '' }); }}
                          className="btn btn-ghost btn-sm btn-row"
                        >
                          <IconSwap className="h-4 w-4" /> Mutasi
                        </button>
                      </div>
                    </td>
                  )}
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {!loading && items.length > 0 && (
          <TableFoot>
            <Pagination page={meta.page} totalPages={meta.total_pages} total={meta.total} noun="aset" onPage={(p) => load(p)} />
          </TableFoot>
        )}
      </div>

      {showForm && (
        <Modal title={editing ? 'Ubah aset' : 'Tambah aset'} subtitle={editing ? 'Perbarui detail aset lalu simpan perubahan.' : 'Kode aset dibuat otomatis setelah disimpan.'} icon={<IconPlus className="h-5 w-5" />} onClose={() => setShowForm(false)} onSubmit={submit} submitLabel={editing ? 'Simpan perubahan' : 'Tambah aset'} error={err} busy={saving}>
          <Field label="Nama aset" htmlFor="f-nama">
            <input id="f-nama" required value={form.nama} onChange={(e) => setForm({ ...form, nama: e.target.value })} placeholder="mis. Laptop Lenovo T14" className="input" />
          </Field>
          <Field label="Serial number" htmlFor="f-sn">
            <input id="f-sn" value={form.serial_number} onChange={(e) => setForm({ ...form, serial_number: e.target.value })} placeholder="Opsional" className="input" />
          </Field>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="Kategori" htmlFor="f-cat">
              <Select id="f-cat" required value={form.category_id} onChange={(v) => setForm({ ...form, category_id: v })}
                placeholder="Pilih kategori" options={cats.map((c) => ({ value: c.id, label: c.nama }))} />
            </Field>
            {!editing && (
              <Field label="Lokasi awal" htmlFor="f-loc">
                <Select id="f-loc" required value={form.location_id} onChange={(v) => setForm({ ...form, location_id: v })}
                  placeholder="Pilih lokasi" options={locs.map((l) => ({ value: l.id, label: l.nama }))} />
              </Field>
            )}
          </div>
          <Field label="Tanggal pembelian" htmlFor="f-date">
            <input id="f-date" type="date" value={form.purchase_date} onChange={(e) => setForm({ ...form, purchase_date: e.target.value })} className="input" />
          </Field>
          <Field label="Catatan" htmlFor="f-note">
            <textarea id="f-note" value={form.catatan} onChange={(e) => setForm({ ...form, catatan: e.target.value })} placeholder="Opsional" className="input" rows={2} />
          </Field>
        </Modal>
      )}

      {moveId && (
        <Modal title="Mutasi aset" subtitle="Pindahkan aset ke lokasi lain dan catat siapa pemegangnya." icon={<IconSwap className="h-5 w-5" />} onClose={() => setMoveId(null)} onSubmit={submitMove} submitLabel="Simpan mutasi" error={err} busy={saving}>
          <Field label="Lokasi tujuan" htmlFor="m-loc">
            <Select id="m-loc" required value={move.to_location_id} onChange={(v) => setMove({ ...move, to_location_id: v })}
              placeholder="Pilih lokasi tujuan" options={locs.map((l) => ({ value: l.id, label: l.nama }))} />
          </Field>
          <Field label="Pemegang" htmlFor="m-holder" hint="ID pemegang (UUID). Kosongkan jika tanpa pemegang.">
            <input id="m-holder" value={move.to_holder_id} onChange={(e) => setMove({ ...move, to_holder_id: e.target.value })} placeholder="UUID pemegang" className="input" />
          </Field>
          <Field label="Catatan mutasi" htmlFor="m-note">
            <textarea id="m-note" value={move.catatan} onChange={(e) => setMove({ ...move, catatan: e.target.value })} placeholder="Opsional" className="input" rows={2} />
          </Field>
        </Modal>
      )}
    </div>
  );
};

export default Assets;
