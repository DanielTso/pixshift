import { Link } from 'react-router';

export default function Footer() {
  return (
    <footer className="border-t border-navy-700/50 bg-navy-950">
      <div className="mx-auto flex max-w-7xl flex-col items-center justify-between gap-4 px-4 py-8 sm:flex-row sm:px-6 lg:px-8">
        <div className="flex items-center gap-2 text-sm text-navy-500">
          <svg width="20" height="20" viewBox="0 0 32 32" fill="none">
            <rect width="32" height="32" rx="6" fill="#0f172a" />
            <path d="M8 12L16 6L24 12V20L16 26L8 20V12Z" stroke="#475569" strokeWidth="2" fill="none" />
            <circle cx="16" cy="16" r="4" fill="#475569" />
          </svg>
          Pixshift &copy; {new Date().getFullYear()}
        </div>
        <nav className="flex items-center gap-6 text-sm text-navy-500">
          <Link to="/docs" className="transition hover:text-navy-300">Docs</Link>
          <Link to="/pricing" className="transition hover:text-navy-300">Pricing</Link>
          <a
            href="https://github.com/pocketbase/pixshift"
            target="_blank"
            rel="noopener noreferrer"
            className="transition hover:text-navy-300"
          >
            GitHub
          </a>
        </nav>
      </div>
    </footer>
  );
}
