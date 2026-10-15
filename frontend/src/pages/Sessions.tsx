import React, { useEffect, useState } from 'react';
import { apiFetch, type ApiError } from '../api';
import { IconLogout } from '../components/Icons';
import { Empty, ErrorAlert, PageHeader, TableBar, TableSkeleton, confirmAction, toast } from '../components/ui';

interface Sess {
  id: string;
  user_id: string;
  expires_at: string;
  revoked_at?: string | null;
  user_agent?: string | null;
  ip_address?: string | null;
  created_at: string;
}

const fmt = (s: string) =>
  new Date(s).toLocaleString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });

const Sessions: React.FC = () => {
  const [items, setItems] = useState<Sess[]>([]);
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  const load = () => {
    apiFetch('/api/v1/auth/sessions')
      .then(setItems)
      .catch((e: ApiError) => setErr(e.message))
      .finally(() => setLoading(false));
  };
  useEffect(() => { load(); }, []);

  const revoke = async (id: string) => {
    if (!(await confirmAction({
      title: 'Cabut sesi ini?',
      body: 'Perangkat ini akan langsung keluar dari akun Anda dan harus masuk ulang.',
      confirmLabel: 'Cabut sesi',
      icon: <IconLogout className="h-5 w-5" />,
      tone: 'danger',
    }))) return;
    try {
      await apiFetch(`/api/v1/auth/sessions/${id}`, { method: 'DELETE' });
      toast('Sesi dicabut');
      load();
    } catch (e) {
      setErr((e as ApiError).message);
    }
  };

  return (
    <div className="space-y-6">
      <PageHeader title="Sesi saya" subtitle="Perangkat yang sedang masuk ke akun Anda. Cabut sesi yang tidak dikenali." />
      {err && <ErrorAlert message={err} />}
      <div className="table-card">
        <TableBar title="Sesi aktif" count={loading ? undefined : items.length} noun="sesi" />
        {loading ? <TableSkeleton rows={3} /> : items.length === 0 ? (
          <Empty text="Tidak ada sesi aktif" hint="Sesi baru akan muncul di sini setiap kali Anda masuk." />
        ) : (
          <table className="rtable">
            <thead>
              <tr>
                <th>Dibuat</th>
                <th>Perangkat</th>
                <th>Alamat IP</th>
                <th>Kedaluwarsa</th>
                <th><span className="sr-only">Aksi</span></th>
              </tr>
            </thead>
            <tbody>
              {items.map((s) => (
                <tr key={s.id}>
                  <td className="cell-main font-medium" data-label="Dibuat">{fmt(s.created_at)}</td>
                  <td className="cell-wide" data-label="Perangkat"><div className="clip" title={s.user_agent ?? ''}>{s.user_agent ?? '—'}</div></td>
                  <td data-label="Alamat IP"><span className="mono">{s.ip_address ?? '—'}</span></td>
                  <td data-label="Kedaluwarsa">{fmt(s.expires_at)}</td>
                  <td className="cell-actions" data-label="Aksi">
                    <div className="row-actions">
                      <button onClick={() => revoke(s.id)} className="btn btn-danger btn-sm btn-row">Cabut sesi</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default Sessions;
