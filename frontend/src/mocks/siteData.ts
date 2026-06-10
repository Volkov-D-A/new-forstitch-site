import { blog } from './blog';
import { categories } from './categories';
import { gallery } from './gallery';
import { products } from './products';
import { author, howToBuy, testimonials } from './siteContent';
import type { SiteData } from '../types/site';

export const siteData: SiteData = {
  author,
  blog,
  categories,
  gallery,
  howToBuy,
  products,
  testimonials,
};
