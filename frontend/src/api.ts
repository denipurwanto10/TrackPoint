// Klien API: access token di memori modul (bukan localStorage).
// Refresh single-flight: request paralel berbagi satu promise refresh.
let accessToken: string | null = null;
let ongoingRefresh: Promise<boolean> | null = null;

export function setToken(t: string | null) {
  accessToken = t;
}
export function getToken() {
  return accessToken;
}

export interface ApiError {
  code: string;
  message: string;
  status: number;
  /** true bila session expired dan perlu login ulang */
  isAuthExpired?: boolean;
}

export class AuthExpiredError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthExpiredError';
  }
}

async function doRefresh(): Promise<boolean> {
  try {
    const r = await fetch('/api/v1/auth/refresh', { method: 'POST', credentials: 'include' });
    if (!r.ok) {
      accessToken = null;
      return false;
    }
    const body = await r.json();
    accessToken = body.data.access_token as string;
    return true;
  } catch {
    accessToken = null;
    return false;
  }
}

export function refreshAccess(): Promise<boolean> {
  if (!ongoingRefresh) {
    ongoingRefresh = doRefresh().finally(() => {
      ongoingRefresh = null;
    });
  }
  return ongoingRefresh;
}

export async function apiFetch(path: string, init: RequestInit = {}, retried = false): Promise<any> {
  const headers: Record<string, string> = { ...(init.headers as Record<string, string> | undefined) };
  if (init.body && !headers['Content-Type']) headers['Content-Type'] = 'application/json';
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
  const resp = await fetch(path, { ...init, headers, credentials: 'include' });
  if (resp.status === 204) return null;
  if (resp.status === 401 && !retried && !path.includes('/auth/refresh') && !path.includes('/auth/login')) {
    const ok = await refreshAccess();
    if (ok) return apiFetch(path, init, true);
    // refresh gagal → session expired
    setToken(null);
    throw new AuthExpiredError('Sesi berakhir, silakan login ulang');
  }
  const body = await resp.json().catch(() => null);
  if (!resp.ok) {
    const err: ApiError = {
      code: body?.error?.code ?? 'REQUEST_GAGAL',
      message: body?.error?.message ?? `Request gagal (${resp.status}).`,
      status: resp.status,
    };
    throw err;
  }
  return body?.data !== undefined ? body.data : body;
}

export async function apiPage(path: string, init: RequestInit = {}): Promise<{ items: any[]; meta: any }> {
  const headers: Record<string, string> = { ...(init.headers as Record<string, string> | undefined) };
  if (accessToken) headers['Authorization'] = `Bearer ${accessToken}`;
  let resp = await fetch(path, { ...init, headers, credentials: 'include' });
  if (resp.status === 401 && !path.includes('/auth/')) {
    const ok = await refreshAccess();
    if (ok) {
      const h2: Record<string, string> = { ...(init.headers as Record<string, string> | undefined) };
      if (accessToken) h2['Authorization'] = `Bearer ${accessToken}`;
      resp = await fetch(path, { ...init, headers: h2, credentials: 'include' });
    } else {
      setToken(null);
      throw new AuthExpiredError('Sesi berakhir, silakan login ulang');
    }
  }
  if (resp.status === 204) return { items: [], meta: {} };
  const body = await resp.json().catch(() => null);
  if (!resp.ok) {
    throw {
      code: body?.error?.code ?? 'REQUEST_GAGAL',
      message: body?.error?.message ?? `Request gagal (${resp.status}).`,
      status: resp.status,
    } as ApiError;
  }
  return { items: body.data ?? [], meta: body.meta ?? {} };
}
