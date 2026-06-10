import { API_BASE_URL } from './siteApi';
import type { Category, Product } from '../types/site';

interface APIErrorPayload {
  error?: {
    code?: string;
    message?: string;
  };
}

export interface AdminSession {
  authenticated: boolean;
  username?: string;
  csrfToken?: string;
}

export interface AdminLoginResponse {
  username: string;
  csrfToken: string;
}

export class AdminAPIError extends Error {
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
    throw new AdminAPIError(
      response.status,
      payload.error?.code || `http_${response.status}`,
      payload.error?.message || response.statusText,
    );
  }

  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}

function csrfHeaders(csrfToken: string): HeadersInit {
  return { 'X-CSRF-Token': csrfToken };
}

async function readError(response: Response): Promise<APIErrorPayload> {
  try {
    return await response.json() as APIErrorPayload;
  } catch {
    return {};
  }
}

export function loginAdmin(username: string, password: string): Promise<AdminLoginResponse> {
  return request<AdminLoginResponse>('/auth/login', {
    method: 'POST',
    body: JSON.stringify({ username, password }),
  });
}

export function getAdminSession(): Promise<AdminSession> {
  return request<AdminSession>('/auth/session');
}

export function logoutAdmin(csrfToken: string): Promise<void> {
  return request<void>('/auth/logout', {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
  });
}

export function getAdminCategories(): Promise<Category[]> {
  return request<Category[]>('/admin/categories');
}

export function createAdminCategory(csrfToken: string, category: Category): Promise<Category> {
  return request<Category>('/admin/categories', {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(category),
  });
}

export function updateAdminCategory(csrfToken: string, category: Category): Promise<Category> {
  return request<Category>(`/admin/categories/${category.id}`, {
    method: 'PUT',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(category),
  });
}

export function deleteAdminCategory(csrfToken: string, categoryId: string): Promise<void> {
  return request<void>(`/admin/categories/${categoryId}`, {
    method: 'DELETE',
    headers: csrfHeaders(csrfToken),
  });
}

export function getAdminProducts(): Promise<Product[]> {
  return request<Product[]>('/admin/products');
}

export function createAdminProduct(csrfToken: string, product: Product): Promise<Product> {
  return request<Product>('/admin/products', {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(product),
  });
}

export function updateAdminProduct(csrfToken: string, product: Product): Promise<Product> {
  return request<Product>(`/admin/products/${product.id}`, {
    method: 'PUT',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(product),
  });
}

export function deleteAdminProduct(csrfToken: string, productId: string): Promise<void> {
  return request<void>(`/admin/products/${productId}`, {
    method: 'DELETE',
    headers: csrfHeaders(csrfToken),
  });
}
