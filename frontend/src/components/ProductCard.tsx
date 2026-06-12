import React from 'react';
import { SImg } from './SImg';
import type { Category, FormatPrice, Product, ProductIdHandler } from '../types/site';

interface ProductImageProps {
  product: Product;
  className?: string;
}

function ProductImage({ product, className }: ProductImageProps) {
  if (product.img) return <SImg src={product.img} alt={product.title} loading="lazy" className={className} />;

  return (
    <div className="ph-img">
      <span className="x-mark">× × ×</span>
      <span>{product.title}</span>
    </div>
  );
}

interface ProductCardProps {
  product: Product;
  categories: Category[];
  formatPrice: FormatPrice;
  onOpen: ProductIdHandler;
  onAdd: ProductIdHandler;
  inCart: boolean;
}

export function ProductCard({ product, categories, formatPrice, onOpen, onAdd, inCart }: ProductCardProps) {
  const category = categories.find((item) => item.id === product.cat);

  return (
    <article className="pcard" onClick={() => onOpen(product.id)}>
      <div className="pcard-imgwrap">
        <ProductImage product={product} />
        {product.isNew ? <span className="pcard-badge">Новинка</span> : null}
      </div>
      <div className="pcard-body">
        <span className="pcard-cat">{category ? category.label : ''}</span>
        <h3 className="pcard-title">{product.title}</h3>
        <div className="pcard-foot">
          <span className="pcard-price">{formatPrice(product.price)}</span>
          <button
            className={'pcard-add' + (inCart ? ' added' : '')}
            title={inCart ? 'В корзине' : 'В корзину'}
            onClick={(event) => {
              event.stopPropagation();
              onAdd(product.id);
            }}
          >
            {inCart ? '✓' : '+'}
          </button>
        </div>
      </div>
    </article>
  );
}
