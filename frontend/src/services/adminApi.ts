import { API_BASE_URL } from './siteApi';
import type { BlogPost, Category, GalleryItem, Product, SiteSettings, Testimonial } from '../types/site';

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
  const isFormData = init?.body instanceof FormData;
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      ...(init?.body && !isFormData ? { 'Content-Type': 'application/json' } : {}),
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

export function getAdminBlog(): Promise<BlogPost[]> {
  return request<BlogPost[]>('/admin/blog');
}

export function getAdminGallery(): Promise<GalleryItem[]> {
  return request<GalleryItem[]>('/admin/gallery');
}

export function createAdminGalleryItem(csrfToken: string, item: GalleryItem): Promise<GalleryItem> {
  return request<GalleryItem>('/admin/gallery', {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(item),
  });
}

export function updateAdminGalleryItem(csrfToken: string, item: GalleryItem): Promise<GalleryItem> {
  return request<GalleryItem>(`/admin/gallery/${item.id}`, {
    method: 'PUT',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(item),
  });
}

export function uploadAdminGalleryItemImage(csrfToken: string, itemId: number, file: File): Promise<GalleryItem> {
  const body = new FormData();
  body.append('file', file);

  return request<GalleryItem>(`/admin/gallery/${itemId}/image`, {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body,
  });
}

export function deleteAdminGalleryItem(csrfToken: string, itemId: number): Promise<void> {
  return request<void>(`/admin/gallery/${itemId}`, {
    method: 'DELETE',
    headers: csrfHeaders(csrfToken),
  });
}

export function createAdminBlogPost(csrfToken: string, post: BlogPost): Promise<BlogPost> {
  return request<BlogPost>('/admin/blog', {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(post),
  });
}

export function updateAdminBlogPost(csrfToken: string, post: BlogPost): Promise<BlogPost> {
  return request<BlogPost>(`/admin/blog/${post.id}`, {
    method: 'PUT',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(post),
  });
}

export function uploadAdminBlogPostImage(csrfToken: string, postId: string, file: File): Promise<BlogPost> {
  const body = new FormData();
  body.append('file', file);

  return request<BlogPost>(`/admin/blog/${postId}/image`, {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body,
  });
}

export function deleteAdminBlogPost(csrfToken: string, postId: string): Promise<void> {
  return request<void>(`/admin/blog/${postId}`, {
    method: 'DELETE',
    headers: csrfHeaders(csrfToken),
  });
}

export function getAdminSiteSettings(): Promise<SiteSettings> {
  return request<SiteSettings>('/admin/site-settings');
}

export function updateAdminSiteSettings(csrfToken: string, settings: SiteSettings): Promise<SiteSettings> {
  return request<SiteSettings>('/admin/site-settings', {
    method: 'PUT',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(settings),
  });
}

export function getAdminTestimonials(): Promise<Testimonial[]> {
  return request<Testimonial[]>('/admin/testimonials');
}

export function createAdminTestimonial(csrfToken: string, testimonial: Testimonial): Promise<Testimonial> {
  return request<Testimonial>('/admin/testimonials', {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(testimonial),
  });
}

export function updateAdminTestimonial(csrfToken: string, testimonial: Testimonial): Promise<Testimonial> {
  return request<Testimonial>(`/admin/testimonials/${testimonial.id}`, {
    method: 'PUT',
    headers: csrfHeaders(csrfToken),
    body: JSON.stringify(testimonial),
  });
}

export function uploadAdminTestimonialImage(csrfToken: string, testimonialId: number, file: File): Promise<Testimonial> {
  const body = new FormData();
  body.append('file', file);

  return request<Testimonial>(`/admin/testimonials/${testimonialId}/image`, {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body,
  });
}

export function deleteAdminTestimonial(csrfToken: string, testimonialId: number): Promise<void> {
  return request<void>(`/admin/testimonials/${testimonialId}`, {
    method: 'DELETE',
    headers: csrfHeaders(csrfToken),
  });
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

export function uploadAdminProductImage(csrfToken: string, productId: string, file: File): Promise<Product> {
  const body = new FormData();
  body.append('file', file);

  return request<Product>(`/admin/products/${productId}/image`, {
    method: 'POST',
    headers: csrfHeaders(csrfToken),
    body,
  });
}

export function deleteAdminProduct(csrfToken: string, productId: string): Promise<void> {
  return request<void>(`/admin/products/${productId}`, {
    method: 'DELETE',
    headers: csrfHeaders(csrfToken),
  });
}
