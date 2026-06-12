import type { CategoryId } from '../types/site';

export const ROUTES = {
  home: '/',
  shop: '/shop',
  account: '/account',
  gallery: '/gallery',
  blog: '/blog',
  howToBuy: '/how-to-buy',
};

export function categoryPath(categoryId: CategoryId = 'all') {
  return categoryId === 'all' ? ROUTES.shop : `${ROUTES.shop}/${categoryId}`;
}

export function productPath(productId: string) {
  return `/product/${productId}`;
}

export function blogPostPath(postId: string) {
  return `${ROUTES.blog}/${postId}`;
}
