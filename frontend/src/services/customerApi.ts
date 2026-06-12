import { API_BASE_URL } from './siteApi';
import type { CustomerOrder, CustomerSession } from '../types/site';

interface APIErrorPayload {
  error?: {
    code?: string;
    message?: string;
  };
}

export class CustomerAPIError extends Error {
  code: string;
  status: number;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  });

  if (!response.ok) {
    const payload = await readError(response);
    throw new CustomerAPIError(
      response.status,
      payload.error?.code || `http_${response.status}`,
      payload.error?.message || response.statusText,
    );
  }

  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

async function readError(response: Response): Promise<APIErrorPayload> {
  try {
    return await response.json() as APIErrorPayload;
  } catch {
    return {};
  }
}

export function loginCustomer(email: string, password: string): Promise<CustomerSession> {
  return request<CustomerSession>('/customer/login', {
    method: 'POST',
    body: JSON.stringify({ username: email, password }),
  });
}

export function startCustomerRegistration(email: string, name: string, password: string): Promise<{ email: string; message: string }> {
  return request<{ email: string; message: string }>('/customer/register/start', {
    method: 'POST',
    body: JSON.stringify({ email, name, password }),
  });
}

export function verifyCustomerRegistration(email: string, code: string): Promise<CustomerSession> {
  return request<CustomerSession>('/customer/register/verify', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  });
}

export function startPasswordReset(email: string): Promise<{ email: string; message: string }> {
  return request<{ email: string; message: string }>('/customer/password-reset/start', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
}

export function verifyPasswordReset(email: string, code: string, newPassword: string): Promise<CustomerSession> {
  return request<CustomerSession>('/customer/password-reset/verify', {
    method: 'POST',
    body: JSON.stringify({ email, code, newPassword }),
  });
}

export function getCustomerSession(): Promise<CustomerSession> {
  return request<CustomerSession>('/customer/session');
}

export function logoutCustomer(): Promise<void> {
  return request<void>('/customer/logout', { method: 'POST' });
}

export async function getCustomerOrders(): Promise<CustomerOrder[]> {
  const orders = await request<CustomerOrder[] | null>('/customer/orders');
  return Array.isArray(orders) ? orders : [];
}
