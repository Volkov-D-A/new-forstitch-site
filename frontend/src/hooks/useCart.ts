import React from 'react';
import type { CartItem, Product } from '../types/site';

interface UseCartOptions {
  products: Product[];
  onAdded?: (product: Product | undefined) => void;
}

export function useCart({ products, onAdded }: UseCartOptions) {
  const [cart, setCart] = React.useState<CartItem[]>([]);
  const [isCartOpen, setCartOpen] = React.useState(false);

  const addToCart = React.useCallback((id: string) => {
    setCart((currentCart) => {
      if (currentCart.some((item) => item.productId === id)) {
        setCartOpen(true);
        return currentCart;
      }

      const product = products.find((item) => item.id === id);
      onAdded?.(product);
      return [...currentCart, { productId: id, quantity: 1 }];
    });
  }, [onAdded, products]);

  const removeFromCart = React.useCallback((id: string) => {
    setCart((currentCart) => currentCart.filter((item) => item.productId !== id));
  }, []);

  const clearCart = React.useCallback(() => {
    setCart([]);
  }, []);

  return {
    addToCart,
    cart,
    cartCount: cart.length,
    clearCart,
    closeCart: () => setCartOpen(false),
    isInCart: (id: string) => cart.some((item) => item.productId === id),
    isCartOpen,
    openCart: () => setCartOpen(true),
    removeFromCart,
  };
}
