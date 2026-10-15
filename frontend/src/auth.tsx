import React, { createContext, useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiFetch, refreshAccess, setToken, AuthExpiredError } from './api';

interface AuthContextType {
  user: { id: string; nama: string; email: string; role: string; is_active: boolean } | null;
  /** true selama sesi lama masih dipulihkan dari cookie refresh (reload halaman). */
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('AuthContext not provided');
  return ctx;
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<AuthContextType['user']>(null);
  // Mulai true: rute tidak boleh memutuskan "belum login" sebelum pemulihan sesi selesai,
  // kalau tidak halaman login sempat berkedip setiap kali halaman di-reload.
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  // api.ts menyimpan access token di memori modul + single-flight refresh via cookie.
  const login = async (email: string, password: string) => {
    const data = await apiFetch('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
    setToken(data.access_token as string);
    setUser(data.user);
    navigate('/');
  };

  const logout = async () => {
    try {
      await apiFetch('/api/v1/auth/logout', { method: 'POST' });
    } catch {
      /* abaikan — tetap bersihkan sesi lokal */
    } finally {
      setToken(null);
    }
    setUser(null);
    navigate('/login');
  };

  // Saat reload: token memori hilang, pulihkan via cookie refresh sekali.
  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        const ok = await refreshAccess();
        if (!ok) return;
        const me = await apiFetch('/api/v1/auth/me');
        if (alive) setUser(me);
      } catch (e) {
        if (e instanceof AuthExpiredError) {
          setToken(null);
        }
        /* tetap di halaman login */
      } finally {
        if (alive) setLoading(false);
      }
    })();
    return () => { alive = false; };
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
