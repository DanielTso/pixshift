import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router';
import { useAuthStore } from '../stores/auth';
import { apiFetch } from '../api/client';

export default function Settings() {
  const user = useAuthStore((s) => s.user);
  const loading = useAuthStore((s) => s.loading);
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [passwordMsg, setPasswordMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState('');

  useEffect(() => {
    if (!loading && !user) {
      navigate('/login');
    }
  }, [user, loading, navigate]);

  if (loading || !user) return null;

  async function handleChangePassword(e: React.FormEvent) {
    e.preventDefault();
    setPasswordMsg(null);
    try {
      await apiFetch('/internal/auth/password', {
        method: 'PUT',
        body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
      });
      setPasswordMsg({ type: 'success', text: 'Password updated successfully.' });
      setCurrentPassword('');
      setNewPassword('');
    } catch (err) {
      setPasswordMsg({ type: 'error', text: (err as Error).message });
    }
  }

  async function handleDeleteAccount() {
    if (deleteConfirm !== user!.email) return;
    setDeleting(true);
    try {
      await apiFetch('/internal/auth/account', { method: 'DELETE' });
      await logout();
      navigate('/');
    } catch {
      setDeleting(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl px-4 py-8 sm:px-6">
      <h1 className="mb-8 text-2xl font-bold text-white">Settings</h1>

      {/* Profile */}
      <section className="mb-8 rounded-xl border border-navy-700/50 bg-navy-800/50 p-6">
        <h2 className="mb-4 text-sm font-semibold text-white">Profile</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-xs text-navy-400">Name</label>
            <p className="text-sm text-navy-200">{user.name}</p>
          </div>
          <div>
            <label className="mb-1 block text-xs text-navy-400">Email</label>
            <p className="text-sm text-navy-200">{user.email}</p>
          </div>
          <div>
            <label className="mb-1 block text-xs text-navy-400">Provider</label>
            <p className="text-sm capitalize text-navy-200">{user.provider}</p>
          </div>
          <div>
            <label className="mb-1 block text-xs text-navy-400">Plan</label>
            <span className={`inline-block rounded-full px-2 py-0.5 text-xs font-semibold uppercase ${
              user.tier === 'pro'
                ? 'bg-accent/15 text-accent'
                : 'bg-navy-700 text-navy-400'
            }`}>
              {user.tier}
            </span>
          </div>
        </div>
      </section>

      {/* Change Password */}
      {user.provider === 'email' && (
        <section className="mb-8 rounded-xl border border-navy-700/50 bg-navy-800/50 p-6">
          <h2 className="mb-4 text-sm font-semibold text-white">Change Password</h2>
          <form onSubmit={handleChangePassword} className="flex flex-col gap-4">
            {passwordMsg && (
              <div className={`rounded-lg px-4 py-2 text-sm ${
                passwordMsg.type === 'success' ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'
              }`}>
                {passwordMsg.text}
              </div>
            )}
            <div>
              <label htmlFor="current-password" className="mb-1 block text-xs text-navy-400">Current Password</label>
              <input
                id="current-password"
                type="password"
                required
                value={currentPassword}
                onChange={(e) => setCurrentPassword(e.target.value)}
                className="w-full rounded-lg border border-navy-600 bg-navy-900 px-3 py-2 text-sm text-white placeholder-navy-500 outline-none focus:border-accent"
              />
            </div>
            <div>
              <label htmlFor="new-password" className="mb-1 block text-xs text-navy-400">New Password</label>
              <input
                id="new-password"
                type="password"
                required
                minLength={8}
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                className="w-full rounded-lg border border-navy-600 bg-navy-900 px-3 py-2 text-sm text-white placeholder-navy-500 outline-none focus:border-accent"
              />
            </div>
            <button
              type="submit"
              className="self-start rounded-lg bg-accent px-4 py-2 text-sm font-medium text-navy-900 transition hover:bg-accent-light"
            >
              Update Password
            </button>
          </form>
        </section>
      )}

      {/* Danger Zone */}
      <section className="rounded-xl border border-red-500/20 bg-red-500/5 p-6">
        <h2 className="mb-2 text-sm font-semibold text-red-400">Danger Zone</h2>
        <p className="mb-4 text-sm text-navy-400">
          Once you delete your account, there is no going back. All data will be permanently removed.
        </p>
        <div className="flex flex-col gap-3">
          <div>
            <label htmlFor="delete-confirm" className="mb-1 block text-xs text-navy-400">
              Type <span className="text-red-400">{user.email}</span> to confirm
            </label>
            <input
              id="delete-confirm"
              type="text"
              value={deleteConfirm}
              onChange={(e) => setDeleteConfirm(e.target.value)}
              className="w-full rounded-lg border border-navy-600 bg-navy-900 px-3 py-2 text-sm text-white placeholder-navy-500 outline-none focus:border-red-500"
              placeholder={user.email}
            />
          </div>
          <button
            onClick={handleDeleteAccount}
            disabled={deleteConfirm !== user.email || deleting}
            className="self-start rounded-lg bg-red-500/20 px-4 py-2 text-sm font-medium text-red-400 transition hover:bg-red-500/30 disabled:opacity-40 disabled:cursor-not-allowed"
          >
            {deleting ? 'Deleting...' : 'Delete Account'}
          </button>
        </div>
      </section>
    </div>
  );
}
