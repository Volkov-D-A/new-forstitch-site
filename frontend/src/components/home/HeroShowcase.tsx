import React from 'react';
import { useNavigate } from 'react-router-dom';
import { EmptyState, SImg, Stitch } from '../index';
import { HOME_FEATURED_PRODUCT_ID } from '../../utils/homeContent';
import { findProduct, firstProductWithImage } from '../../utils/products';
import { categoryPath, productPath, ROUTES } from '../../utils/routes';
import type { Category, FormatPrice, Product, SiteData } from '../../types/site';

interface CategoryTilesProps {
  categories: Category[];
  products: Product[];
}

function CategoryTiles({ categories, products }: CategoryTilesProps) {
  const navigate = useNavigate();

  return (
    <section className="sec home-category-section" data-screen-label="Главная: категории">
      <div className="wrap">
        <div className="sec-head home-section-head-compact">
          <div>
            <Stitch />
            <h2 className="h-sec">Каталог по темам</h2>
          </div>
        </div>
        <div className="cat-tiles">
          {categories.filter((category) => category.id !== 'all').map((category) => {
            const count = products.filter((product) => product.cat === category.id).length;
            const preview = firstProductWithImage(products, (product) => product.cat === category.id);

            return (
              <div className="cat-tile" key={category.id} onClick={() => navigate(categoryPath(category.id))}>
                {preview
                  ? <SImg src={preview.img} alt={category.label} loading="lazy" />
                  : <div className="ph-img"><span className="x-mark">×××</span></div>}
                <div className="cat-tile-label"><span>{category.label}</span><span>{count} схем</span></div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}

interface HeroShowcaseProps {
  data: SiteData;
  formatPrice: FormatPrice;
}

export function HeroShowcase({ data, formatPrice }: HeroShowcaseProps) {
  const navigate = useNavigate();
  const hero = findProduct(data.products, data.featuredProductId || HOME_FEATURED_PRODUCT_ID);

  if (!hero) {
    return <EmptyState title="Витринный товар не найден" text="Проверьте настройки главного товара в данных сайта." />;
  }

  return (
    <React.Fragment>
      <section className="heroC" data-screen-label="Главная: герой (Витрина)">
        <div className="heroC-bg"><SImg src={hero.img} alt="" /></div>
        <div className="wrap heroC-content">
          <p className="eyebrow heroC-eyebrow">Новинка месяца</p>
          <h1 className="h-display">{hero.title}</h1>
          <p className="lede">
            58 цветов, градиенты неба и моря — морская серия открыта.
            Схема в PDF, доставка на почту сразу после оплаты.
          </p>
          <div className="heroA-cta">
            <button className="btn btn-primary" onClick={() => navigate(productPath(hero.id))}>Купить за {formatPrice(hero.price)}</button>
            <button className="btn btn-outline heroC-outline" onClick={() => navigate(ROUTES.shop)}>Весь каталог</button>
          </div>
        </div>
      </section>
      <CategoryTiles categories={data.categories} products={data.products} />
    </React.Fragment>
  );
}
