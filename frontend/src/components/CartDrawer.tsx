import React from 'react';
import { SImg } from './SImg';
import type { FormatPrice, Product, ProductIdHandler } from '../types/site';

interface CartDrawerProps {
  cart: string[];
  onClose: () => void;
  onRemove: ProductIdHandler;
  onShopOpen: () => void;
  products: Product[];
  formatPrice: FormatPrice;
}

export function CartDrawer({ cart, onClose, onRemove, onShopOpen, products, formatPrice }: CartDrawerProps) {
  const items = cart
    .map((id) => products.find((product) => product.id === id))
    .filter((product): product is Product => Boolean(product));
  const total = items.reduce((sum, product) => sum + product.price, 0);

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
          ) : items.map((product) => (
            <div className="cart-row" key={product.id}>
              {product.img
                ? <SImg className="cart-thumb" src={product.img} alt={product.title} />
                : <div className="cart-thumb ph-img"><span className="x-mark tiny">×××</span></div>}
              <div className="cart-info">
                <div className="cart-title">{product.title}</div>
                <div className="cart-price">PDF-схема · {formatPrice(product.price)}</div>
              </div>
              <button className="cart-x" title="Убрать" onClick={() => onRemove(product.id)}>✕</button>
            </div>
          ))}
        </div>
        {items.length > 0 ? (
          <div className="drawer-foot">
            <div className="cart-total"><span>Итого</span><span>{formatPrice(total)}</span></div>
            <button className="btn btn-primary drawer-checkout">Оформить заказ</button>
            <p className="drawer-note">
              Схемы придут на почту сразу после оплаты
            </p>
          </div>
        ) : null}
      </aside>
    </React.Fragment>
  );
}
