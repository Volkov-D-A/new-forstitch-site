import React from 'react';
import { SImg } from './SImg';
import type { CartItem, FormatPrice, Product, ProductIdHandler } from '../types/site';

interface CartDrawerProps {
  cart: CartItem[];
  isCheckoutLoading: boolean;
  onClose: () => void;
  onCheckout: () => void;
  onQuantityChange: (productId: string, quantity: number) => void;
  onRemove: ProductIdHandler;
  onShopOpen: () => void;
  products: Product[];
  formatPrice: FormatPrice;
}

interface CartDrawerItem {
  product: Product;
  quantity: number;
}

export function CartDrawer({
  cart,
  formatPrice,
  isCheckoutLoading,
  onCheckout,
  onClose,
  onQuantityChange,
  onRemove,
  onShopOpen,
  products,
}: CartDrawerProps) {
  const items = cart
    .map((item) => {
      const product = products.find((candidate) => candidate.id === item.productId);
      return product ? { product, quantity: item.quantity } : null;
    })
    .filter((item): item is CartDrawerItem => Boolean(item));
  const total = items.reduce((sum, item) => sum + item.product.price * item.quantity, 0);

  return (
    <React.Fragment>
      <div className="drawer-veil" onClick={onClose}></div>
      <aside className="drawer">
        <div className="drawer-head">
          <h3 className="h-sec drawer-title">Корзина</h3>
          <button className="icon-btn" onClick={onClose}>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="5" y1="5" x2="19" y2="19"></line><line x1="19" y1="5" x2="5" y2="19"></line></svg>
          </button>
        </div>
        <div className="drawer-body">
          {items.length === 0 ? (
            <div className="cart-empty">
              <span className="x-mark accent-lg">× × ×</span>
              <p>В корзине пока пусто</p>
              <button className="btn btn-outline btn-sm" onClick={onShopOpen}>Перейти в магазин</button>
            </div>
          ) : items.map(({ product, quantity }) => (
            <div className="cart-row" key={product.id}>
              {product.img
                ? <SImg className="cart-thumb" src={product.img} alt={product.title} />
                : <div className="cart-thumb ph-img"><span className="x-mark tiny">×××</span></div>}
              <div className="cart-info">
                <div className="cart-title">{product.title}</div>
                <div className="cart-price">PDF-схема · {formatPrice(product.price)}</div>
                <div className="cart-qty">
                  <button onClick={() => onQuantityChange(product.id, quantity - 1)} aria-label="Уменьшить количество">−</button>
                  <span>{quantity}</span>
                  <button onClick={() => onQuantityChange(product.id, quantity + 1)} aria-label="Увеличить количество">+</button>
                </div>
              </div>
              <button className="cart-x" title="Убрать" onClick={() => onRemove(product.id)}>✕</button>
            </div>
          ))}
        </div>
        {items.length > 0 ? (
          <div className="drawer-foot">
            <div className="cart-total"><span>Итого</span><span>{formatPrice(total)}</span></div>
            <button className="btn btn-primary drawer-checkout" disabled={isCheckoutLoading} onClick={onCheckout}>
              {isCheckoutLoading ? 'Оформляем...' : 'Оформить заказ'}
            </button>
            <p className="drawer-note">
              Схемы придут на почту сразу после оплаты
            </p>
          </div>
        ) : null}
      </aside>
    </React.Fragment>
  );
}
