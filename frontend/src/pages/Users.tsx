import React, { useEffect, useState } from 'react';
import { apiFetch, apiPage, type ApiError } from '../api';
import { useAuth } from '../auth';
import { IconPlus } from '../components/Icons';
import { Empty, ErrorAlert, Field, Modal, PageHeader, Select, StatusBadge, TableBar, TableSkeleton, initials, toast } from '../components/ui';

interface U {
  id: string; nama: string; email: string; role: string; is_active: boolean;
}

const Users: React.FC = () => {
  const { user: me } = useAuth();
  const [items, setItems] = useState<U[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [show, setShow] = useState(false);
  const [form, setForm] = useState({ nama: '', email: '', password: '', role: 'KARYAWAN' });

  const load = async () => {
    try {
      const { items } = await apiPage('/api/v1/users?per_page=100');
      setItems(items);
      setErr(null);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(); }, []);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setErr(null);
    try {
      await apiFetch('/api/v1/users', { method: 'POST', body: JSON.stringify(form) });
      setShow(false);
      setForm({ nama: '', email: '', password: '', role: 'KARYAWAN' });
      toast('Pengguna ditambahkan');
      load();
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setSaving(false);
    }
  };

  const toggle = async (u: U) => {
    if (u.id === me?.id) {
      setErr('Tidak bisa menonaktifkan akun sendiri.');
      return;
    }
    try {
      await apiFetch(`/api/v1/users/${u.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ nama: u.nama, role: u.role, is_active: !u.is_active }),
      });
      toast(u.is_active ? 'Pengguna dinonaktifkan' : 'Pengguna diaktifkan');
      load();
    } catch (e) {
      setErr((e as ApiError).message);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Pengguna" subtitle="Kelola akun dan peran akses." />
      {err && !show && <ErrorAlert message={err} />}

      <div className="table-card">
        <TableBar
          title="Daftar pengguna"
          count={loading ? undefined : items.length}
          noun="pengguna"
          action={
            <button onClick={() => { setErr(null); setShow(true); }} className="btn btn-primary">
              <IconPlus className="h-4 w-4" /> Tambah pengguna
            </button>
          }
        />
        {loading ? <TableSkeleton /> : items.length === 0 ? (
          <Empty text="Belum ada pengguna" hint="Tambahkan akun untuk admin, teknisi, atau karyawan." />
        ) : (
          <table className="rtable">
            <thead>
              <tr>
                <th>Pengguna</th>
                <th>Peran</th>
                <th>Status</th>
                <th><span className="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody>
              {items.map((u) => (
                <tr key={u.id}>
                  <td className="cell-main" data-label="Pengguna">
                    <div className="flex items-center gap-3">
                      <span className="avatar">{initials(u.nama)}</span>
                      <div className="min-w-0">
                        <div className="truncate font-medium">{u.nama}{u.id === me?.id && <span className="meta ml-2 font-normal">Anda</span>}</div>
                        <div className="meta truncate">{u.email}</div>
                      </div>
                    </div>
                  </td>
                  <td data-label="Peran"><StatusBadge value={u.role} /></td>
                  <td data-label="Status">
                    <span className={`badge ${u.is_active ? 'badge-ok' : 'badge-mute'}`}>{u.is_active ? 'Aktif' : 'Nonaktif'}</span>
                  </td>
                  <td className="cell-actions" data-label="Aksi">
                    <div className="row-actions">
                      <button onClick={() => toggle(u)} className="btn btn-ghost btn-sm btn-row">
                        {u.is_active ? 'Nonaktifkan' : 'Aktifkan'}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {show && (
        <Modal title="Tambah pengguna" subtitle="Akun baru bisa langsung dipakai untuk masuk." icon={<IconPlus className="h-5 w-5" />} onClose={() => setShow(false)} onSubmit={submit} submitLabel="Tambah pengguna" error={err} busy={saving}>
          <Field label="Nama" htmlFor="u-nama">
            <input id="u-nama" required value={form.nama} onChange={(e) => setForm({ ...form, nama: e.target.value })} placeholder="Nama lengkap" className="input" autoComplete="off" />
          </Field>
          <Field label="Email" htmlFor="u-email">
            <input id="u-email" required type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="nama@perusahaan.com" className="input" autoComplete="off" />
          </Field>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <Field label="Password" htmlFor="u-pw" hint="Minimal 8 karakter.">
              <input id="u-pw" required type="password" minLength={8} value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} placeholder="Password awal" className="input" autoComplete="new-password" />
            </Field>
            <Field label="Peran" htmlFor="u-role">
              <Select id="u-role" value={form.role} onChange={(v) => setForm({ ...form, role: v })} options={[
                { value: 'ADMIN', label: 'Admin' },
                { value: 'TEKNISI', label: 'Teknisi' },
                { value: 'KARYAWAN', label: 'Karyawan' },
              ]} />
            </Field>
          </div>
        </Modal>
      )}
    </div>
  );
};

export default Users;
