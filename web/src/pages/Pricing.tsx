import { Link } from 'react-router';
import { useAuthStore } from '../stores/auth';
import { createCheckout } from '../api/billing';
import { useState } from 'react';

const features = [
  { name: 'Image conversions', free: '5 / day', pro: 'Unlimited' },
  { name: 'Max file size', free: '10 MB', pro: '50 MB' },
  { name: 'Output formats', free: 'All 8 formats', pro: 'All 8 formats' },
  { name: 'Transform options', free: 'All', pro: 'All' },
  { name: 'API access', free: '100 req / day', pro: '10,000 req / day' },
  { name: 'MCP integration', free: '-', pro: 'Included' },
  { name: 'Priority processing', free: '-', pro: 'Included' },
  { name: 'Batch uploads', free: '-', pro: 'Up to 20 files' },
];

const faqs = [
  {
    q: 'Can I convert images without signing up?',
    a: 'Yes! Anonymous users get 5 free conversions per day. Sign up to track your history and get API access.',
  },
  {
    q: 'What formats are supported?',
    a: 'Pixshift supports JPEG, PNG, WebP, AVIF, HEIC, TIFF, GIF (including animated), and JPEG XL.',
  },
  {
    q: 'Are my images stored on your servers?',
    a: 'No. Images are processed in memory and immediately discarded after conversion. We never store your files.',
  },
  {
    q: 'Can I cancel my Pro subscription?',
    a: 'Yes, you can cancel anytime from your Dashboard. You\'ll retain Pro access until the end of your billing period.',
  },
];

export default function Pricing() {
  const user = useAuthStore((s) => s.user);
  const [upgrading, setUpgrading] = useState(false);

  async function handleUpgrade() {
    setUpgrading(true);
    try {
      const url = await createCheckout();
      window.location.href = url;
    } catch {
      setUpgrading(false);
    }
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="mb-12 text-center">
        <h1 className="mb-3 text-4xl font-bold text-white">Simple, transparent pricing</h1>
        <p className="text-lg text-navy-400">Start free. Upgrade when you need more.</p>
      </div>

      {/* Pricing cards */}
      <div className="mx-auto mb-16 grid max-w-3xl gap-6 md:grid-cols-2">
        {/* Free tier */}
        <div className="rounded-2xl border border-navy-700/50 bg-navy-800/50 p-8">
          <h2 className="mb-1 text-lg font-semibold text-white">Free</h2>
          <p className="mb-6 text-sm text-navy-400">For occasional use</p>
          <div className="mb-6">
            <span className="text-4xl font-bold text-white">$0</span>
            <span className="text-navy-400"> / month</span>
          </div>
          <Link
            to={user ? '/dashboard' : '/signup'}
            className="mb-8 block rounded-xl border border-navy-600 py-2.5 text-center text-sm font-medium text-navy-300 transition hover:border-navy-500 hover:text-white"
          >
            {user ? 'Go to Dashboard' : 'Get Started'}
          </Link>
          <ul className="flex flex-col gap-3 text-sm text-navy-300">
            {features.map((f) => (
              <li key={f.name} className="flex items-center justify-between">
                <span>{f.name}</span>
                <span className="text-navy-400">{f.free}</span>
              </li>
            ))}
          </ul>
        </div>

        {/* Pro tier */}
        <div className="rounded-2xl border border-accent/30 bg-navy-800/50 p-8 shadow-[0_0_30px_rgba(6,182,212,0.05)]">
          <div className="mb-1 flex items-center gap-2">
            <h2 className="text-lg font-semibold text-white">Pro</h2>
            <span className="rounded-full bg-accent/15 px-2 py-0.5 text-[10px] font-semibold uppercase text-accent">
              Popular
            </span>
          </div>
          <p className="mb-6 text-sm text-navy-400">For professionals and teams</p>
          <div className="mb-6">
            <span className="text-4xl font-bold text-white">$9</span>
            <span className="text-navy-400"> / month</span>
          </div>
          {user?.tier === 'pro' ? (
            <div className="mb-8 rounded-xl bg-accent/10 py-2.5 text-center text-sm font-medium text-accent">
              Current Plan
            </div>
          ) : (
            <button
              onClick={user ? handleUpgrade : undefined}
              disabled={upgrading}
              className="mb-8 block w-full rounded-xl bg-accent py-2.5 text-center text-sm font-semibold text-navy-900 transition hover:bg-accent-light disabled:opacity-50"
            >
              {!user ? (
                <Link to="/signup" className="block">Sign up for Pro</Link>
              ) : upgrading ? 'Redirecting...' : 'Upgrade to Pro'}
            </button>
          )}
          <ul className="flex flex-col gap-3 text-sm text-navy-300">
            {features.map((f) => (
              <li key={f.name} className="flex items-center justify-between">
                <span>{f.name}</span>
                <span className="font-medium text-accent">{f.pro}</span>
              </li>
            ))}
          </ul>
        </div>
      </div>

      {/* FAQ */}
      <div className="mx-auto max-w-2xl">
        <h2 className="mb-8 text-center text-2xl font-bold text-white">Frequently Asked Questions</h2>
        <div className="flex flex-col gap-4">
          {faqs.map((faq) => (
            <div key={faq.q} className="rounded-xl border border-navy-700/50 bg-navy-800/50 p-5">
              <h3 className="mb-2 text-sm font-semibold text-white">{faq.q}</h3>
              <p className="text-sm text-navy-400">{faq.a}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
