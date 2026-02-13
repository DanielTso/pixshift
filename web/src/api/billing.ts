import { apiFetch } from './client';

interface CheckoutResponse {
  url: string;
}

interface PortalResponse {
  url: string;
}

export async function createCheckout(plan: 'pro' | 'business', interval: 'monthly' | 'annual'): Promise<string> {
  const res = await apiFetch<CheckoutResponse>('/internal/billing/checkout', {
    method: 'POST',
    body: JSON.stringify({ plan, interval }),
  });
  return res.url;
}

export async function createPortal(): Promise<string> {
  const res = await apiFetch<PortalResponse>('/internal/billing/portal', {
    method: 'POST',
  });
  return res.url;
}
