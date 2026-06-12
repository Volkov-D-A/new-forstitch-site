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
  img?: string;
  images?: ProductImage[];
  files?: ProductFile[];
  ph?: string;
  isNew?: boolean;
  size: string;
  colors: string;
  description?: string;
}

export interface ProductImage {
  id: number;
  url: string;
}

export interface ProductFile {
  id: number;
  name: string;
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
  description: string;
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
  status?: string;
}

export interface OrderItem {
  productId: string;
  productName: string;
  quantity: number;
  price: number;
  downloads?: DownloadFile[];
}

export interface DownloadFile {
  id: number;
  name: string;
  url: string;
}

export interface CustomerOrder {
  id: string;
  status: string;
  customerEmail: string;
  customerName?: string;
  message?: string;
  items: OrderItem[];
  createdAt: string;
}

export interface CustomerSession {
  authenticated: boolean;
  email?: string;
  name?: string;
}

export type FormatPrice = (price: number) => string;
export type ProductIdHandler = (productId: string) => void;
