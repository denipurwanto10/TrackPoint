import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './auth';
import Login from './Login';
import { Logo } from './components/Icons';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Assets from './pages/Assets';
import Maintenance from './pages/Maintenance';
import Master from './pages/Master';
import Users from './pages/Users';
import Sessions from './pages/Sessions';

/** Layar jeda singkat saat sesi lama masih dipulihkan (refresh cookie) — mencegah kedipan halaman login. */
const Booting: React.FC = () => (
  <div className="boot" role="status" aria-live="polite">
    <Logo size={40} />
    <span className="spinner" aria-hidden="true" />
    <span className="boot-text">Menyiapkan sesi…</span>
  </div>
);

function Guard({ children, roles }: { children: React.ReactElement; roles?: string[] }) {
  const { user, loading } = useAuth();
  if (loading) return <Booting />;
  if (!user) return <Navigate to="/login" replace />;
  if (roles && !roles.includes(user.role)) return <Navigate to="/" replace />;
  return children;
}

const App: React.FC = () => {
  const { user, loading } = useAuth();
  return (
    <Routes>
      <Route path="/login" element={loading ? <Booting /> : user ? <Navigate to="/" replace /> : <Login />} />
      <Route element={<Guard><Layout /></Guard>}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/aset" element={<Assets />} />
        <Route path="/maintenance" element={<Maintenance />} />
        <Route path="/sesi" element={<Sessions />} />
        <Route path="/master" element={<Guard roles={['ADMIN']}><Master /></Guard>} />
        <Route path="/pengguna" element={<Guard roles={['ADMIN']}><Users /></Guard>} />
      </Route>
      <Route path="*" element={loading ? <Booting /> : <Navigate to={user ? '/' : '/login'} replace />} />
    </Routes>
  );
};

export default App;
