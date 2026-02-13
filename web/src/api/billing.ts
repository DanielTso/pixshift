import { apiFetch } from './client';

interface CheckoutResponse {
  url: string;
}

interface PortalResponse {
  url: string;
}

export async function createCheckout(): Promise<string> {
  const res = await apiFetch<CheckoutResponse>('/internal/billing/checkout', {
    method: 'POST',
  });
  return res.url;
}

export async function createPortal(): Promise<string> {
  const res = await apiFetch<PortalResponse>('/internal/billing/portal', {
    method: 'POST',
  });
  return res.url;
}
