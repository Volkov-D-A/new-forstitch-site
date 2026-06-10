import React from 'react';
import { SImg, Stitch } from '../components/index';
import type { GalleryItem, SiteData } from '../types/site';

interface GalleryItemCardProps {
  item: GalleryItem;
  onOpen: (item: GalleryItem) => void;
}

function GalleryItemCard({ item, onOpen }: GalleryItemCardProps) {
  return (
    <div className="masonry-item" onClick={() => onOpen(item)}>
      <SImg src={item.img} alt={item.title} loading="lazy" />
      <div className="masonry-cap"><b>{item.title}</b><span>{item.by}</span></div>
    </div>
  );
}

interface LightboxProps {
  item: GalleryItem | null;
  onClose: () => void;
}

function Lightbox({ item, onClose }: LightboxProps) {
  if (!item) return null;

  return (
    <div className="lightbox" onClick={onClose}>
      <SImg src={item.img} alt={item.title} />
      <div className="lightbox-cap">{item.title} — {item.by}</div>
    </div>
  );
}

interface GalleryPageProps {
  data: SiteData;
}

export function GalleryPage({ data }: GalleryPageProps) {
  const [openItem, setOpenItem] = React.useState<GalleryItem | null>(null);

  React.useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setOpenItem(null);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);

  return (
    <div data-screen-label="Галерея">
      <div className="wrap page-head">
        <Stitch />
        <h1 className="h-sec page-title">Галерея отшивов</h1>
        <p className="lede page-lede">
          Работы, вышитые рукодельницами по схемам Екатерины. Пришлите свой отшив — и он появится здесь.
        </p>
      </div>
      <div className="wrap page-content">
        <div className="masonry">
          {data.gallery.map((item, index) => (
            <GalleryItemCard key={item.id || index} item={item} onOpen={setOpenItem} />
          ))}
        </div>
      </div>
      <Lightbox item={openItem} onClose={() => setOpenItem(null)} />
    </div>
  );
}
