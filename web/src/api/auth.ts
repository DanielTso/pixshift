import { apiFetch } from './client';

export interface User {
  id: string;
  email: string;
  name: string;
  provider: 'email' | 'google';
  tier: 'free' | 'pro' | 'business';
  created_at: string;
}

export async function login(email: string, password: string): Promise<User> {
  return apiFetch<User>('/internal/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
}

export async function signup(email: string, password: string, name: string): Promise<User> {
  return apiFetch<User>('/internal/auth/signup', {
    method: 'POST',
    body: JSON.stringify({ email, password, name }),
  });
}

export async function logout(): Promise<void> {
  await apiFetch<void>('/internal/auth/logout', { method: 'POST' });
}

export async function getUser(): Promise<User> {
  return apiFetch<User>('/internal/auth/me');
}
