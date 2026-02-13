import { useEffect } from 'react';
import { useNavigate, Link } from 'react-router';
import { useAuthStore } from '../stores/auth';
import UsageChart from '../components/dashboard/UsageChart';
import ConversionHistory from '../components/dashboard/ConversionHistory';
import APIKeyManager from '../components/dashboard/APIKeyManager';

export default function Dashboard() {
  const user = useAuthStore((s) => s.user);
  const loading = useAuthStore((s) => s.loading);
  const navigate = useNavigate();

  useEffect(() => {
    if (!loading && !user) {
      navigate('/login');
    }
  }, [user, loading, navigate]);

  if (loading) {
    return (
      <div className="flex min-h-[calc(100vh-12rem)] items-center justify-center">
        <p className="text-navy-500">Loading...</p>
      </div>
    );
  }

  if (!user) return null;

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="flex flex-col gap-8 lg:flex-row">
        {/* Sidebar */}
        <aside className="w-full lg:w-56 shrink-0">
          <div className="rounded-xl border border-navy-700/50 bg-navy-800/50 p-4">
            <div className="mb-4 flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-accent text-sm font-bold text-navy-900">
                {user.name.charAt(0).toUpperCase()}
              </div>
              <div>
                <p className="text-sm font-medium text-white">{user.name}</p>
                <span className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase ${
                  user.tier === 'business'
                    ? 'bg-amber-400/15 text-amber-400'
                    : user.tier === 'pro'
                      ? 'bg-accent/15 text-accent'
                      : 'bg-navy-700 text-navy-400'
                }`}>
                  {user.tier}
                </span>
              </div>
            </div>
            <nav className="flex flex-col gap-1">
              <Link to="/dashboard" className="rounded-lg bg-navy-700/50 px-3 py-2 text-sm text-white">
                Overview
              </Link>
              <Link to="/settings" className="rounded-lg px-3 py-2 text-sm text-navy-400 hover:bg-navy-700/30 hover:text-white">
                Settings
              </Link>
              {user.tier === 'free' && (
                <Link to="/pricing" className="mt-2 rounded-lg bg-accent/10 px-3 py-2 text-center text-sm font-medium text-accent hover:bg-accent/15">
                  Upgrade
                </Link>
              )}
              {user.tier === 'pro' && (
                <Link to="/pricing" className="mt-2 rounded-lg bg-amber-400/10 px-3 py-2 text-center text-sm font-medium text-amber-400 hover:bg-amber-400/15">
                  Upgrade to Business
                </Link>
              )}
            </nav>
          </div>
        </aside>

        {/* Main content */}
        <div className="flex flex-1 flex-col gap-6">
          <div>
            <h1 className="text-2xl font-bold text-white">Dashboard</h1>
            <p className="text-sm text-navy-400">Monitor your conversions and manage API keys.</p>
          </div>
          <UsageChart />
          <ConversionHistory />
          <APIKeyManager />
        </div>
      </div>
    </div>
  );
}
