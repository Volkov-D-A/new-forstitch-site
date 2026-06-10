import React from 'react';
import { useNavigate } from 'react-router-dom';
import { ProductCard, Stitch } from '../index';
import { productPath, ROUTES } from '../../utils/routes';
import type { FormatPrice, ProductIdHandler, SiteData } from '../../types/site';

interface HomeNewSectionProps {
  addToCart: ProductIdHandler;
  isInCart: (productId: string) => boolean;
  data: SiteData;
  formatPrice: FormatPrice;
}

export function HomeNewSection({ addToCart, isInCart, data, formatPrice }: HomeNewSectionProps) {
  const navigate = useNavigate();
  const products = data.products.filter((product) => product.isNew);

  return (
    <section className="sec" data-screen-label="Главная: новинки">
      <div className="wrap">
        <div className="sec-head">
          <div>
            <Stitch />
            <h2 className="h-sec">Новинки магазина</h2>
          </div>
          <button className="btn btn-ghost" onClick={() => navigate(ROUTES.shop)}>Все схемы →</button>
        </div>
        <div className="pgrid">
          {products.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              categories={data.categories}
              formatPrice={formatPrice}
              onOpen={(id: string) => navigate(productPath(id))}
              onAdd={addToCart}
              inCart={isInCart(product.id)}
            />
          ))}
        </div>
      </div>
    </section>
  );
}
