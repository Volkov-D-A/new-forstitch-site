import React from 'react';
import { useNavigate } from 'react-router-dom';
import { EmptyState, SImg } from '../index';
import { HOME_FEATURED_PRODUCT_ID } from '../../utils/homeContent';
import { findProduct } from '../../utils/products';
import { productPath, ROUTES } from '../../utils/routes';
import type { FormatPrice, SiteData } from '../../types/site';

interface HeroCozyProps {
  data: SiteData;
  formatPrice: FormatPrice;
}

export function HeroCozy({ data, formatPrice }: HeroCozyProps) {
  const navigate = useNavigate();
  const featured = findProduct(data.products, data.featuredProductId || HOME_FEATURED_PRODUCT_ID);

  if (!featured) {
    return <EmptyState title="Товар для главного экрана не найден" text="Проверьте настройки главного товара в данных сайта." />;
  }

  return (
    <section className="heroA canvas-bg" data-screen-label="Главная: герой (Уют)">
      <div className="wrap heroA-grid">
        <div>
          <p className="eyebrow">Схемы для вышивки крестом ручной разработки</p>
          <h1 className="h-display">Живопись, перенесённая в&nbsp;крестики</h1>
          <p className="lede">
            Авторские схемы Екатерины Волковой: плавные переходы цветов, объём
            и натуралистичность, за которые их любят даже искушённые вышивальщицы.
          </p>
          <div className="heroA-cta">
            <button className="btn btn-primary" onClick={() => navigate(ROUTES.shop)}>Выбрать схему</button>
            <button className="btn btn-outline" onClick={() => navigate(ROUTES.gallery)}>Смотреть отшивы</button>
          </div>
          <div className="heroA-stats">
            <div className="heroA-stat"><b>150+</b><span>авторских схем</span></div>
            <div className="heroA-stat"><b>15 лет</b><span>разработки схем</span></div>
            <div className="heroA-stat"><b>PDF</b><span>мгновенная доставка</span></div>
          </div>
        </div>
        <div className="heroA-imgwrap">
          <div className="heroA-stitchframe"></div>
          <div className="heroA-img" onClick={() => navigate(productPath(featured.id))}>
            <SImg src={featured.img} alt={featured.title} />
          </div>
          <div className="heroA-tag" onClick={() => navigate(productPath(featured.id))}>
            <b>{featured.title}</b>
            <span>новинка · {formatPrice(featured.price)}</span>
          </div>
        </div>
      </div>
    </section>
  );
}
