export type CategoryId = 'all' | 'fauna' | 'people' | 'still-life' | 'landscape' | 'fantasy';

export interface Category {
  id: CategoryId;
  label: string;
}

export interface Product {
  id: string;
  title: string;
  price: number;
  cat: Exclude<CategoryId, 'all'>;
  sub: string;
  img?: string;
  ph?: string;
  isNew?: boolean;
  size: string;
  colors: string;
  canvas: string;
}

export interface Testimonial {
  name: string;
  role: string;
  img: string;
  text: string;
}

export interface GalleryItem {
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
  howToBuy: HowToStep[];
}

export type FormatPrice = (price: number) => string;
export type ProductIdHandler = (productId: string) => void;
