import type { Product } from '../types/site';

export function findProduct(products: Product[], id: string): Product | undefined {
  return products.find((product) => product.id === id);
}

export function firstProductWithImage(
  products: Product[],
  predicate: (product: Product) => boolean = () => true,
): Product | undefined {
  return products.find((product) => predicate(product) && product.img);
}
