import type {
  BlogPost,
  Category,
  GalleryItem,
  OrderRequest,
  OrderResponse,
  Product,
  SiteData,
} from '../types/site';

export const API_BASE_URL = 'http://localhost:3000/api';

type SiteContentResponse = Pick<SiteData, 'author' | 'howToBuy' | 'testimonials'>;

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init?.body ? { 'Content-Type': 'application/json' } : {}),
      ...init?.headers,
    },
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status} ${response.statusText}`);
  }

  return response.json() as Promise<T>;
}

function normalizeCategory(category: Category): Category {
  return {
    id: String(category.id),
    label: category.label || String(category.id),
  };
}

function normalizeProduct(product: Product): Product {
  return {
    ...product,
    id: String(product.id),
    title: product.title || 'Схема без названия',
    price: Number(product.price) || 0,
    cat: String(product.cat),
    sub: product.sub || '',
    img: product.img || undefined,
    isNew: Boolean(product.isNew),
    size: product.size || '',
    colors: product.colors || '',
    canvas: product.canvas || '',
  };
}

export async function getCategories(): Promise<Category[]> {
  const categories = await request<Category[]>('/categories');
  const normalized = categories.map(normalizeCategory);
  return normalized.some((category) => category.id === 'all')
    ? normalized
    : [{ id: 'all', label: 'Все схемы' }, ...normalized];
}

export async function getProducts(): Promise<Product[]> {
  const products = await request<Product[]>('/products');
  return products.map(normalizeProduct);
}

export async function getProduct(productId: string): Promise<Product | null> {
  try {
    return normalizeProduct(await request<Product>(`/products/${productId}`));
  } catch {
    return null;
  }
}

export async function getGallery(): Promise<GalleryItem[]> {
  return request<GalleryItem[]>('/gallery');
}

export async function getBlog(): Promise<BlogPost[]> {
  return request<BlogPost[]>('/blog');
}

export async function getSiteContent(): Promise<SiteContentResponse> {
  return request<SiteContentResponse>('/site-content');
}

export async function getSiteData(): Promise<SiteData> {
  const [categories, products, gallery, blog, siteContent] = await Promise.all([
    getCategories(),
    getProducts(),
    getGallery(),
    getBlog(),
    getSiteContent(),
  ]);

  return {
    author: siteContent.author,
    blog,
    categories,
    gallery,
    howToBuy: siteContent.howToBuy,
    products,
    testimonials: siteContent.testimonials,
  };
}

export async function createOrder(order: OrderRequest): Promise<OrderResponse> {
  return request<OrderResponse>('/orders', {
    method: 'POST',
    body: JSON.stringify(order),
  });
}
