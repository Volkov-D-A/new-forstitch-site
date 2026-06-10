import { blog as mockBlog } from '../mocks/blog';
import { categories as mockCategories } from '../mocks/categories';
import { gallery as mockGallery } from '../mocks/gallery';
import { products as mockProducts } from '../mocks/products';
import { author as mockAuthor, howToBuy as mockHowToBuy, testimonials as mockTestimonials } from '../mocks/siteContent';
import type { BlogPost, Category, GalleryItem, Product, SiteData } from '../types/site';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

async function request<T>(path: string): Promise<T | null> {
  if (!API_BASE_URL) return null;

  const response = await fetch(`${API_BASE_URL}${path}`, {
    headers: { Accept: 'application/json' },
  });

  if (!response.ok) {
    throw new Error(`API request failed: ${response.status} ${response.statusText}`);
  }

  return response.json() as Promise<T>;
}

export async function getCategories(): Promise<Category[]> {
  return (await request<Category[]>('/categories')) || mockCategories;
}

export async function getProducts(): Promise<Product[]> {
  return (await request<Product[]>('/products')) || mockProducts;
}

export async function getProduct(productId: string): Promise<Product | null> {
  const remoteProduct = await request<Product>(`/products/${productId}`);
  if (remoteProduct) return remoteProduct;
  return mockProducts.find((product) => product.id === productId) || null;
}

export async function getGallery(): Promise<GalleryItem[]> {
  return (await request<GalleryItem[]>('/gallery')) || mockGallery;
}

export async function getBlog(): Promise<BlogPost[]> {
  return (await request<BlogPost[]>('/blog')) || mockBlog;
}

export async function getSiteData(): Promise<SiteData> {
  const [categories, products, gallery, blog] = await Promise.all([
    getCategories(),
    getProducts(),
    getGallery(),
    getBlog(),
  ]);

  return {
    author: mockAuthor,
    blog,
    categories,
    gallery,
    howToBuy: mockHowToBuy,
    products,
    testimonials: mockTestimonials,
  };
}
