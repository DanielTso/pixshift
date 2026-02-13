import { Link } from 'react-router';
import { useAuthStore } from '../stores/auth';
import { createCheckout } from '../api/billing';
import { useState } from 'react';

type Interval = 'monthly' | 'annual';

const prices = {
  pro: { monthly: 19, annual: 190 },
  business: { monthly: 59, annual: 590 },
};

const features = [
  { name: 'Web Conversions', starter: '20/day', pro: '500/day', business: 'Unlimited' },
  { name: 'API Requests', starter: '100/month', pro: '5,000/month', business: '50,000/month' },
  { name: 'Max File Size', starter: '10 MB', pro: '100 MB', business: '500 MB' },
  { name: 'Output Formats', starter: 'All 8 formats', pro: 'All 8 formats', business: 'All 8 formats' },
  { name: 'Transform Tools', starter: 'All', pro: 'All', business: 'All' },
  { name: 'Batch Uploads', starter: '1 file', pro: '20 files', business: '100 files' },
  { name: 'MCP Integration', starter: '\u2014', pro: 'Included', business: 'Included' },
  { name: 'API Keys', starter: '1', pro: '5', business: '20' },
  { name: 'Rate Limit', starter: '10 req/min', pro: '30 req/min', business: '120 req/min' },
  { name: 'Support', starter: 'Community', pro: 'Priority Email', business: 'Dedicated' },
];

const faqs = [
  {
    q: 'Can I convert images without signing up?',
    a: 'Yes! Anonymous users get 20 free web conversions per day. Sign up to track your history and get API access.',
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
    q: 'Can I cancel my subscription?',
    a: 'Yes, you can cancel anytime from your Dashboard. You\'ll retain access until the end of your billing period.',
  },
  {
    q: 'Do you offer annual billing?',
    a: 'Yes! Save ~17% with annual billing \u2014 equivalent to 2 months free.',
  },
  {
    q: 'Can I switch plans?',
    a: 'Yes, you can upgrade or downgrade anytime. Changes take effect immediately, with prorated billing.',
  },
];

export default function Pricing() {
  const user = useAuthStore((s) => s.user);
  const [interval, setInterval] = useState<Interval>('monthly');
  const [upgradingPlan, setUpgradingPlan] = useState<'pro' | 'business' | null>(null);

  async function handleUpgrade(plan: 'pro' | 'business') {
    setUpgradingPlan(plan);
    try {
      const url = await createCheckout(plan, interval);
      window.location.href = url;
    } catch {
      setUpgradingPlan(null);
    }
  }

  function formatAnnualMonthly(annual: number) {
    return `$${(annual / 12).toFixed(2)}/mo billed annually`;
  }

  return (
    <div className="mx-auto max-w-6xl px-4 py-16 sm:px-6 lg:px-8">
      <div className="mb-12 text-center">
        <h1 className="mb-3 text-4xl font-bold text-white">Simple, transparent pricing</h1>
        <p className="text-lg text-navy-400">Start free. Upgrade when you need more.</p>
      </div>

      {/* Interval toggle */}
      <div className="mb-12 flex items-center justify-center gap-3">
        <button
          onClick={() => setInterval('monthly')}
          className={`rounded-xl px-4 py-2 text-sm font-medium transition ${
            interval === 'monthly'
              ? 'bg-navy-700 text-white'
              : 'text-navy-400 hover:text-white'
          }`}
        >
          Monthly
        </button>
        <button
          onClick={() => setInterval('annual')}
          className={`flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-medium transition ${
            interval === 'annual'
              ? 'bg-navy-700 text-white'
              : 'text-navy-400 hover:text-white'
          }`}
        >
          Annual
          <span className="rounded-full bg-accent/15 px-2 py-0.5 text-[10px] font-semibold text-accent">
            2 months free
          </span>
        </button>
      </div>

      {/* Pricing cards */}
      <div className="mx-auto mb-16 grid max-w-5xl gap-6 md:grid-cols-3">
        {/* Starter tier */}
        <div className="rounded-2xl border border-navy-700/50 bg-navy-800/50 p-8">
          <h2 className="mb-1 text-lg font-semibold text-white">Starter</h2>
          <p className="mb-6 text-sm text-navy-400">For personal use</p>
          <div className="mb-2">
            <span className="text-4xl font-bold text-white">$0</span>
            <span className="text-navy-400"> / month</span>
          </div>
          <p className="mb-6 text-sm text-navy-500">&nbsp;</p>
          {user?.tier === 'free' ? (
            <div className="mb-8 rounded-xl bg-navy-700/50 py-2.5 text-center text-sm font-medium text-navy-300">
              Current Plan
            </div>
          ) : (
            <Link
              to={user ? '/dashboard' : '/signup'}
              className="mb-8 block rounded-xl border border-navy-600 py-2.5 text-center text-sm font-medium text-navy-300 transition hover:border-navy-500 hover:text-white"
            >
              {user ? 'Go to Dashboard' : 'Get Started'}
            </Link>
          )}
          <ul className="flex flex-col gap-3 text-sm text-navy-300">
            {features.map((f) => (
              <li key={f.name} className="flex items-center justify-between">
                <span>{f.name}</span>
                <span className="text-navy-400">{f.starter}</span>
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
          <p className="mb-6 text-sm text-navy-400">For developers</p>
          <div className="mb-2">
            <span className="text-4xl font-bold text-white">
              ${interval === 'monthly' ? prices.pro.monthly : prices.pro.annual}
            </span>
            <span className="text-navy-400">
              {interval === 'monthly' ? ' / month' : ' / year'}
            </span>
          </div>
          <p className="mb-6 text-sm text-navy-500">
            {interval === 'annual' ? formatAnnualMonthly(prices.pro.annual) : '\u00A0'}
          </p>
          {user?.tier === 'pro' ? (
            <div className="mb-8 rounded-xl bg-accent/10 py-2.5 text-center text-sm font-medium text-accent">
              Current Plan
            </div>
          ) : (
            <button
              onClick={user ? () => handleUpgrade('pro') : undefined}
              disabled={upgradingPlan === 'pro'}
              className="mb-8 block w-full rounded-xl bg-accent py-2.5 text-center text-sm font-semibold text-navy-900 transition hover:bg-accent-light disabled:opacity-50"
            >
              {!user ? (
                <Link to="/signup" className="block">Sign up for Pro</Link>
              ) : upgradingPlan === 'pro' ? 'Redirecting...' : 'Upgrade to Pro'}
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

        {/* Business tier */}
        <div className="rounded-2xl border border-amber-400/30 bg-navy-800/50 p-8 shadow-[0_0_30px_rgba(251,191,36,0.05)]">
          <div className="mb-1 flex items-center gap-2">
            <h2 className="text-lg font-semibold text-white">Business</h2>
            <span className="rounded-full bg-amber-400/15 px-2 py-0.5 text-[10px] font-semibold uppercase text-amber-400">
              Teams
            </span>
          </div>
          <p className="mb-6 text-sm text-navy-400">For teams & enterprises</p>
          <div className="mb-2">
            <span className="text-4xl font-bold text-white">
              ${interval === 'monthly' ? prices.business.monthly : prices.business.annual}
            </span>
            <span className="text-navy-400">
              {interval === 'monthly' ? ' / month' : ' / year'}
            </span>
          </div>
          <p className="mb-6 text-sm text-navy-500">
            {interval === 'annual' ? formatAnnualMonthly(prices.business.annual) : '\u00A0'}
          </p>
          {user?.tier === 'business' ? (
            <div className="mb-8 rounded-xl bg-amber-400/10 py-2.5 text-center text-sm font-medium text-amber-400">
              Current Plan
            </div>
          ) : (
            <button
              onClick={user ? () => handleUpgrade('business') : undefined}
              disabled={upgradingPlan === 'business'}
              className="mb-8 block w-full rounded-xl bg-amber-400 py-2.5 text-center text-sm font-semibold text-navy-900 transition hover:bg-amber-300 disabled:opacity-50"
            >
              {!user ? (
                <Link to="/signup" className="block text-navy-900">Sign up for Business</Link>
              ) : upgradingPlan === 'business' ? 'Redirecting...' : 'Upgrade to Business'}
            </button>
          )}
          <ul className="flex flex-col gap-3 text-sm text-navy-300">
            {features.map((f) => (
              <li key={f.name} className="flex items-center justify-between">
                <span>{f.name}</span>
                <span className="font-medium text-amber-400">{f.business}</span>
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
