export type CategoryId = string;

export interface Category {
  id: CategoryId;
  label: string;
}

export interface Product {
  id: string;
  title: string;
  price: number;
  cat: CategoryId;
  sub: string;
  img?: string;
  ph?: string;
  isNew?: boolean;
  size: string;
  colors: string;
  canvas: string;
}

export interface Testimonial {
  id?: number;
  name: string;
  role: string;
  img: string;
  text: string;
}

export interface GalleryItem {
  id?: number;
  img: string;
  title: string;
  by: string;
  w?: number;
  tall?: boolean;
}

export interface Author {
  name: string;
  photo: string;
  p1: string;
  p2: string;
  p3: string;
  sign: string;
}

export interface BlogPost {
  id: string;
  title: string;
  date: string;
  tag: string;
  img: string;
  excerpt: string;
  content: string;
}

export interface HowToStep {
  n: string;
  t: string;
  d: string;
}

export interface SiteData {
  categories: Category[];
  products: Product[];
  testimonials: Testimonial[];
  gallery: GalleryItem[];
  author: Author;
  blog: BlogPost[];
  featuredProductId?: string;
  howToBuy: HowToStep[];
}

export interface SiteSettings {
  featuredProductId: string;
}

export interface CartItem {
  productId: string;
  quantity: number;
}

export interface OrderRequest {
  items: CartItem[];
}

export interface OrderResponse {
  id: string;
  checkoutUrl?: string;
  message?: string;
}

export type FormatPrice = (price: number) => string;
export type ProductIdHandler = (productId: string) => void;
