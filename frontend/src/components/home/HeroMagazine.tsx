import React from 'react';
import { useNavigate } from 'react-router-dom';
import { SImg } from '../index';
import { HOME_MAGAZINE_CAPTIONS, HOME_MAGAZINE_PRODUCT_IDS } from '../../utils/homeContent';
import { findProduct } from '../../utils/products';
import { productPath, ROUTES } from '../../utils/routes';
import type { Product, SiteData } from '../../types/site';

interface HeroMagazineProps {
  data: SiteData;
}

export function HeroMagazine({ data }: HeroMagazineProps) {
  const navigate = useNavigate();
  const marquee = 'Натюрморты × Пейзажи × Фэнтези × Животный мир × Люди × ';
  const collageProducts = HOME_MAGAZINE_PRODUCT_IDS
    .map((id) => findProduct(data.products, id))
    .filter((product): product is Product => Boolean(product));

  return (
    <section className="heroB" data-screen-label="Главная: герой (Журнал)">
      <div className="wrap">
        <h1 className="h-display">
          Каждый крестик —<br />на <span className="accent-i">своём</span> месте
        </h1>
        <div className="heroB-row">
          <p className="lede">
            Авторские схемы для вышивки крестом Екатерины Волковой — по живописи,
            фотографии и собственным сюжетам.
          </p>
          <button className="btn btn-primary" onClick={() => navigate(ROUTES.shop)}>В магазин</button>
        </div>
        <div className="heroB-collage">
          {collageProducts.map((product, index) => (
            <div className={`cg cg${index + 1}`} key={product.id} onClick={() => navigate(productPath(product.id))}>
              <SImg src={product.img} alt={product.title} />
              <span className="cg-cap">{HOME_MAGAZINE_CAPTIONS[product.id as keyof typeof HOME_MAGAZINE_CAPTIONS] || product.title}</span>
            </div>
          ))}
        </div>
      </div>
      <div className="heroB-marquee">
        <span className="heroB-marquee-in">
          {[0, 1].map((key) => (
            <span key={key}>{marquee.split('×').filter(Boolean).map((word, index) => (
              <span key={index}>{word}<span className="x-mark">×</span></span>
            ))}</span>
          ))}
        </span>
      </div>
    </section>
  );
}
