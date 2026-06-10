import React from 'react';
import type { Product } from '../types/site';

interface UseCartOptions {
  products: Product[];
  onAdded?: (product: Product | undefined) => void;
}

export function useCart({ products, onAdded }: UseCartOptions) {
  const [cart, setCart] = React.useState<string[]>([]);
  const [isCartOpen, setCartOpen] = React.useState(false);

  const addToCart = React.useCallback((id: string) => {
    setCart((currentCart) => {
      if (currentCart.includes(id)) {
        setCartOpen(true);
        return currentCart;
      }

      const product = products.find((item) => item.id === id);
      onAdded?.(product);
      return [...currentCart, id];
    });
  }, [onAdded, products]);

  const removeFromCart = React.useCallback((id: string) => {
    setCart((currentCart) => currentCart.filter((item) => item !== id));
  }, []);

  return {
    addToCart,
    cart,
    closeCart: () => setCartOpen(false),
    isCartOpen,
    openCart: () => setCartOpen(true),
    removeFromCart,
  };
}
