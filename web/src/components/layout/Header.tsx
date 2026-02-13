import { Link } from 'react-router';
import { useAuthStore } from '../../stores/auth';
import { useState } from 'react';

export default function Header() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <header className="sticky top-0 z-50 border-b border-navy-700/50 bg-navy-900/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-8">
          <Link to="/" className="flex items-center gap-2 text-xl font-bold text-white">
            <svg width="28" height="28" viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="6" fill="#0f172a" />
              <path d="M8 12L16 6L24 12V20L16 26L8 20V12Z" stroke="#06b6d4" strokeWidth="2" fill="none" />
              <circle cx="16" cy="16" r="4" fill="#06b6d4" />
            </svg>
            Pixshift
          </Link>
          <nav className="hidden items-center gap-6 md:flex">
            <Link to="/" className="text-sm text-navy-400 transition hover:text-white">
              Home
            </Link>
            <Link to="/pricing" className="text-sm text-navy-400 transition hover:text-white">
              Pricing
            </Link>
            <Link to="/docs" className="text-sm text-navy-400 transition hover:text-white">
              Docs
            </Link>
          </nav>
        </div>

        <div className="hidden items-center gap-4 md:flex">
          {user ? (
            <div className="relative">
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm text-navy-300 transition hover:bg-navy-800 hover:text-white"
              >
                <div className="flex h-7 w-7 items-center justify-center rounded-full bg-accent text-xs font-bold text-navy-900">
                  {user.name.charAt(0).toUpperCase()}
                </div>
                {user.name}
              </button>
              {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 rounded-lg border border-navy-700 bg-navy-800 py-1 shadow-xl">
                  <Link
                    to="/dashboard"
                    onClick={() => setMenuOpen(false)}
                    className="block px-4 py-2 text-sm text-navy-300 hover:bg-navy-700 hover:text-white"
                  >
                    Dashboard
                  </Link>
                  <Link
                    to="/settings"
                    onClick={() => setMenuOpen(false)}
                    className="block px-4 py-2 text-sm text-navy-300 hover:bg-navy-700 hover:text-white"
                  >
                    Settings
                  </Link>
                  <button
                    onClick={() => { logout(); setMenuOpen(false); }}
                    className="block w-full px-4 py-2 text-left text-sm text-navy-300 hover:bg-navy-700 hover:text-white"
                  >
                    Sign out
                  </button>
                </div>
              )}
            </div>
          ) : (
            <>
              <Link
                to="/login"
                className="rounded-lg px-4 py-2 text-sm text-navy-300 transition hover:text-white"
              >
                Log in
              </Link>
              <Link
                to="/signup"
                className="rounded-lg bg-accent px-4 py-2 text-sm font-medium text-navy-900 transition hover:bg-accent-light"
              >
                Sign up
              </Link>
            </>
          )}
        </div>

        <button
          className="md:hidden text-navy-400 hover:text-white"
          onClick={() => setMenuOpen(!menuOpen)}
        >
          <svg width="24" height="24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
      </div>

      {menuOpen && (
        <div className="border-t border-navy-700/50 bg-navy-900/95 backdrop-blur-xl md:hidden">
          <nav className="flex flex-col px-4 py-3 gap-2">
            <Link to="/" onClick={() => setMenuOpen(false)} className="rounded-lg px-3 py-2 text-sm text-navy-300 hover:bg-navy-800 hover:text-white">Home</Link>
            <Link to="/pricing" onClick={() => setMenuOpen(false)} className="rounded-lg px-3 py-2 text-sm text-navy-300 hover:bg-navy-800 hover:text-white">Pricing</Link>
            <Link to="/docs" onClick={() => setMenuOpen(false)} className="rounded-lg px-3 py-2 text-sm text-navy-300 hover:bg-navy-800 hover:text-white">Docs</Link>
            {user ? (
              <>
                <Link to="/dashboard" onClick={() => setMenuOpen(false)} className="rounded-lg px-3 py-2 text-sm text-navy-300 hover:bg-navy-800 hover:text-white">Dashboard</Link>
                <button onClick={() => { logout(); setMenuOpen(false); }} className="rounded-lg px-3 py-2 text-left text-sm text-navy-300 hover:bg-navy-800 hover:text-white">Sign out</button>
              </>
            ) : (
              <>
                <Link to="/login" onClick={() => setMenuOpen(false)} className="rounded-lg px-3 py-2 text-sm text-navy-300 hover:bg-navy-800 hover:text-white">Log in</Link>
                <Link to="/signup" onClick={() => setMenuOpen(false)} className="rounded-lg px-3 py-2 text-sm text-accent hover:bg-navy-800">Sign up</Link>
              </>
            )}
          </nav>
        </div>
      )}
    </header>
  );
}
