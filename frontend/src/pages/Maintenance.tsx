import React, { useEffect, useState } from 'react';
import { apiFetch, apiPage, type ApiError } from '../api';
import { useAuth } from '../auth';
import { IconPlus, IconWrench } from '../components/Icons';
import { Empty, ErrorAlert, Field, Modal, PageHeader, FilterTabs, Pagination, Select, StatusBadge, TableBar, TableFoot, TableSkeleton, statusLabel, toast } from '../components/ui';

interface Ticket {
  id: string; asset_id: string; asset_code: string; asset_nama: string;
  status: string; keluhan: string;
  tindakan?: string | null; biaya?: number | null; return_to?: string | null;
  reported_by: string; reporter_nama: string;
  technician_id?: string | null; technician_nama?: string | null;
  cancel_reason?: string | null;
  created_at: string;
}

const STATUSES = ['', 'REPORTED', 'IN_PROGRESS', 'DONE', 'CANCELLED'];

const Maintenance: React.FC = () => {
  const { user } = useAuth();
  const canProcess = user?.role === 'ADMIN' || user?.role === 'TEKNISI';
  const [items, setItems] = useState<Ticket[]>([]);
  const [meta, setMeta] = useState({ page: 1, total_pages: 0, total: 0 });
  const [status, setStatus] = useState('');
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showReport, setShowReport] = useState(false);
  const [report, setReport] = useState({ asset_id: '', keluhan: '' });
  const [proc, setProc] = useState<Ticket | null>(null);
  const [to, setTo] = useState('IN_PROGRESS');
  const [fields, setFields] = useState({ tindakan: '', biaya: '', return_to: 'AVAILABLE', cancel_reason: '' });

  const load = async (page = 1, s = status) => {
    try {
      const params = new URLSearchParams({ page: String(page), per_page: '20' });
      if (s) params.set('status', s);
      const { items, meta } = await apiPage(`/api/v1/maintenance?${params}`);
      setItems(items);
      setMeta({ page: meta.page, total_pages: meta.total_pages, total: meta.total });
      setErr(null);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setLoading(false);
    }
  };
  useEffect(() => { load(1); /* eslint-disable-next-line react-hooks/exhaustive-deps */ }, []);

  const submitReport = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setErr(null);
    try {
      await apiFetch('/api/v1/maintenance', { method: 'POST', body: JSON.stringify(report) });
      setShowReport(false);
      setReport({ asset_id: '', keluhan: '' });
      toast('Laporan terkirim');
      load(1);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setSaving(false);
    }
  };

  const submitTransition = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!proc) return;
    setSaving(true);
    setErr(null);
    try {
      await apiFetch(`/api/v1/maintenance/${proc.id}`, {
        method: 'PATCH',
        body: JSON.stringify({
          to,
          tindakan: fields.tindakan || undefined,
          biaya: fields.biaya ? Number(fields.biaya) : undefined,
          return_to: to === 'DONE' ? fields.return_to : undefined,
          cancel_reason: to === 'CANCELLED' ? fields.cancel_reason : undefined,
        }),
      });
      setProc(null);
      toast('Tiket diperbarui');
      load(meta.page);
    } catch (e) {
      setErr((e as ApiError).message);
    } finally {
      setSaving(false);
    }
  };

  const modalOpen = showReport || !!proc;

  return (
    <div className="space-y-6">
      <PageHeader
        title="Maintenance"
        subtitle="Pantau tiket kerusakan dari dilaporkan sampai selesai."
      />
      {err && !modalOpen && <ErrorAlert message={err} />}

      <div className="table-card">
        <TableBar
          title="Daftar tiket"
          count={loading ? undefined : meta.total}
          noun="tiket"
          action={
            <button onClick={() => { setErr(null); setShowReport(true); }} className="btn btn-primary">
              <IconPlus className="h-4 w-4" /> Lapor kerusakan
            </button>
          }
        >
          <FilterTabs label="Filter status" value={status} onChange={(s) => { setStatus(s); load(1, s); }}
            options={STATUSES.map((s) => ({ value: s, label: s ? statusLabel(s) : 'Semua' }))} />
        </TableBar>
        {loading ? <TableSkeleton /> : items.length === 0 ? (
          <Empty
            text={status ? 'Tidak ada tiket dengan status ini' : 'Belum ada tiket'}
            hint={status ? 'Pilih status lain untuk melihat tiket.' : 'Laporkan kerusakan perangkat agar teknisi bisa menanganinya.'}
          />
        ) : (
          <table className="rtable">
            <thead>
              <tr>
                <th>Aset</th>
                <th>Keluhan</th>
                <th>Pelapor</th>
                <th>Teknisi</th>
                <th>Status</th>
                {canProcess && <th><span className="sr-only">Aksi</span></th>}
              </tr>
            </thead>
            <tbody>
              {items.map((t) => (
                <tr key={t.id}>
                  <td className="cell-main" data-label="Aset">
                    <div className="font-medium">{t.asset_nama}</div>
                    <span className="code mt-1 inline-block">{t.asset_code}</span>
                  </td>
                  <td className="cell-wide" data-label="Keluhan"><div className="clip" title={t.keluhan}>{t.keluhan}</div></td>
                  <td data-label="Pelapor">{t.reporter_nama}</td>
                  <td data-label="Teknisi" style={t.technician_nama ? undefined : { color: 'var(--muted)' }}>{t.technician_nama ?? '—'}</td>
                  <td data-label="Status"><StatusBadge value={t.status} /></td>
                  {canProcess && (
                    <td className="cell-actions" data-label="Aksi">
                      <div className="row-actions">
                        {(t.status === 'REPORTED' || t.status === 'IN_PROGRESS') && (
                          <button
                            onClick={() => { setErr(null); setProc(t); setTo(t.status === 'REPORTED' ? 'IN_PROGRESS' : 'DONE'); setFields({ tindakan: '', biaya: '', return_to: 'AVAILABLE', cancel_reason: '' }); }}
                            className="btn btn-ghost btn-sm btn-row"
                          >
                            <IconWrench className="h-4 w-4" /> Proses
                          </button>
                        )}
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
            <Pagination page={meta.page} totalPages={meta.total_pages} total={meta.total} noun="tiket" onPage={(p) => load(p)} />
          </TableFoot>
        )}
      </div>

      {showReport && (
        <Modal title="Lapor kerusakan" subtitle="Catat kerusakan perangkat agar teknisi bisa menindaklanjuti." icon={<IconWrench className="h-5 w-5" />} onClose={() => setShowReport(false)} onSubmit={submitReport} submitLabel="Kirim laporan" error={err} busy={saving}>
          <Field label="ID aset" htmlFor="r-asset">
            <input id="r-asset" required value={report.asset_id} onChange={(e) => setReport({ ...report, asset_id: e.target.value })} placeholder="UUID aset" className="input" />
          </Field>
          <Field label="Keluhan" htmlFor="r-keluhan" hint="Karyawan hanya bisa melaporkan aset yang sedang dipegangnya.">
            <textarea id="r-keluhan" required value={report.keluhan} onChange={(e) => setReport({ ...report, keluhan: e.target.value })} placeholder="Jelaskan masalahnya" className="input" rows={3} />
          </Field>
        </Modal>
      )}

      {proc && (
        <Modal title={`Proses ${proc.asset_code}`} subtitle="Perbarui status tiket dan catat tindakan perbaikannya." icon={<IconWrench className="h-5 w-5" />} onClose={() => setProc(null)} onSubmit={submitTransition} submitLabel="Perbarui tiket" error={err} busy={saving}>
          <Field label="Ubah status ke" htmlFor="p-to">
            <Select id="p-to" value={to} onChange={setTo} options={[
              ...(proc.status === 'REPORTED' ? [{ value: 'IN_PROGRESS', label: 'Dikerjakan' }] : []),
              ...(proc.status === 'IN_PROGRESS' ? [{ value: 'DONE', label: 'Selesai' }] : []),
              { value: 'CANCELLED', label: 'Dibatalkan' },
            ]} />
          </Field>
          {to === 'DONE' && (
            <>
              <Field label="Tindakan perbaikan" htmlFor="p-tindakan">
                <textarea id="p-tindakan" required value={fields.tindakan} onChange={(e) => setFields({ ...fields, tindakan: e.target.value })} placeholder="Apa yang dikerjakan" className="input" rows={2} />
              </Field>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <Field label="Biaya" htmlFor="p-biaya">
                  <input id="p-biaya" type="number" min="0" step="0.01" inputMode="decimal" value={fields.biaya} onChange={(e) => setFields({ ...fields, biaya: e.target.value })} placeholder="Opsional" className="input" />
                </Field>
                <Field label="Kembalikan sebagai" htmlFor="p-return">
                  <Select id="p-return" value={fields.return_to} onChange={(v) => setFields({ ...fields, return_to: v })} options={[
                    { value: 'AVAILABLE', label: 'Tersedia' },
                    { value: 'ASSIGNED', label: 'Dipakai (ke teknisi)' },
                    { value: 'INACTIVE', label: 'Nonaktif' },
                  ]} />
                </Field>
              </div>
            </>
          )}
          {to === 'CANCELLED' && (
            <Field label="Alasan pembatalan" htmlFor="p-cancel">
              <textarea id="p-cancel" required value={fields.cancel_reason} onChange={(e) => setFields({ ...fields, cancel_reason: e.target.value })} placeholder="Kenapa dibatalkan" className="input" rows={2} />
            </Field>
          )}
        </Modal>
      )}
    </div>
  );
};

export default Maintenance;
