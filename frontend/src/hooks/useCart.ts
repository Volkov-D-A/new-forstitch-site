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
        return currentCart.map((item) => (
          item.productId === id ? { ...item, quantity: item.quantity + 1 } : item
        ));
      }

      const product = products.find((item) => item.id === id);
      onAdded?.(product);
      return [...currentCart, { productId: id, quantity: 1 }];
    });
  }, [onAdded, products]);

  const removeFromCart = React.useCallback((id: string) => {
    setCart((currentCart) => currentCart.filter((item) => item.productId !== id));
  }, []);

  const setQuantity = React.useCallback((id: string, quantity: number) => {
    setCart((currentCart) => {
      if (quantity < 1) return currentCart.filter((item) => item.productId !== id);
      return currentCart.map((item) => (
        item.productId === id ? { ...item, quantity } : item
      ));
    });
  }, []);

  const clearCart = React.useCallback(() => {
    setCart([]);
  }, []);

  return {
    addToCart,
    cart,
    cartCount: cart.reduce((sum, item) => sum + item.quantity, 0),
    clearCart,
    closeCart: () => setCartOpen(false),
    isInCart: (id: string) => cart.some((item) => item.productId === id),
    isCartOpen,
    openCart: () => setCartOpen(true),
    removeFromCart,
    setQuantity,
  };
}
