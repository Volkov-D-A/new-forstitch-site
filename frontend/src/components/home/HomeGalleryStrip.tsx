import React from 'react';
import { useNavigate } from 'react-router-dom';
import { SImg, Stitch } from '../index';
import { ROUTES } from '../../utils/routes';
import type { GalleryItem } from '../../types/site';

interface HomeGalleryStripProps {
  gallery: GalleryItem[];
}

export function HomeGalleryStrip({ gallery }: HomeGalleryStripProps) {
  const navigate = useNavigate();
  const items = gallery.slice(0, 4);

  return (
    <section className="sec" data-screen-label="Главная: галерея">
      <div className="wrap">
        <div className="sec-head">
          <div>
            <Stitch />
            <h2 className="h-sec">Отшивы рукодельниц</h2>
          </div>
          <button className="btn btn-ghost" onClick={() => navigate(ROUTES.gallery)}>Вся галерея →</button>
        </div>
        <div className="pgrid home-gallery-grid">
          {items.map((item, index) => (
            <div className="masonry-item home-gallery-card" key={index} onClick={() => navigate(ROUTES.gallery)}>
              <SImg className="home-gallery-image" src={item.img} alt={item.title} loading="lazy" />
              <div className="masonry-cap"><b>{item.title}</b><span>{item.by}</span></div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
