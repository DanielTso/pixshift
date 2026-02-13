import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { useAuthStore } from '../../stores/auth';
import OAuthButton from './OAuthButton';

export default function LoginForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const login = useAuthStore((s) => s.login);
  const error = useAuthStore((s) => s.error);
  const loading = useAuthStore((s) => s.loading);
  const clearError = useAuthStore((s) => s.clearError);
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    try {
      await login(email, password);
      navigate('/dashboard');
    } catch {
      // error is set in store
    }
  }

  return (
    <div className="mx-auto w-full max-w-sm">
      <h1 className="mb-2 text-2xl font-bold text-white">Welcome back</h1>
      <p className="mb-8 text-sm text-navy-400">Sign in to your Pixshift account</p>

      <OAuthButton mode="login" />

      <div className="my-6 flex items-center gap-3">
        <div className="h-px flex-1 bg-navy-700" />
        <span className="text-xs text-navy-500">or</span>
        <div className="h-px flex-1 bg-navy-700" />
      </div>

      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        {error && (
          <div className="rounded-lg bg-red-500/10 px-4 py-2 text-sm text-red-400">
            {error}
            <button onClick={clearError} className="ml-2 text-red-300 hover:text-red-200">&times;</button>
          </div>
        )}
        <div>
          <label htmlFor="email" className="mb-1 block text-sm text-navy-400">Email</label>
          <input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full rounded-lg border border-navy-600 bg-navy-800 px-3 py-2 text-sm text-white placeholder-navy-500 outline-none transition focus:border-accent focus:ring-1 focus:ring-accent"
            placeholder="you@example.com"
          />
        </div>
        <div>
          <label htmlFor="password" className="mb-1 block text-sm text-navy-400">Password</label>
          <input
            id="password"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full rounded-lg border border-navy-600 bg-navy-800 px-3 py-2 text-sm text-white placeholder-navy-500 outline-none transition focus:border-accent focus:ring-1 focus:ring-accent"
            placeholder="Enter your password"
          />
        </div>
        <button
          type="submit"
          disabled={loading}
          className="rounded-lg bg-accent px-4 py-2.5 text-sm font-medium text-navy-900 transition hover:bg-accent-light disabled:opacity-50"
        >
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </form>

      <p className="mt-6 text-center text-sm text-navy-400">
        Don't have an account?{' '}
        <Link to="/signup" className="text-accent hover:text-accent-light">Sign up</Link>
      </p>
    </div>
  );
}
