import React from 'react';
import {
  HeroCozy,
  HeroMagazine,
  HeroShowcase,
  HomeAuthorSection,
  HomeGalleryStrip,
  HomeNewSection,
  HomeTestimonialsSection,
} from '../components/home/index';
import type { FormatPrice, ProductIdHandler, SiteData } from '../types/site';

const HERO_BY_VARIANT = {
  'Уют': HeroCozy,
  'Журнал': HeroMagazine,
  'Витрина': HeroShowcase,
};

type HomeVariant = keyof typeof HERO_BY_VARIANT;

interface HomePageProps {
  variant: HomeVariant;
  data: SiteData;
  formatPrice: FormatPrice;
  addToCart: ProductIdHandler;
  isInCart: (productId: string) => boolean;
}

export function HomePage({ variant, data, formatPrice, addToCart, isInCart }: HomePageProps) {
  const Hero = HERO_BY_VARIANT[variant] || HeroCozy;
  const commonProps = { data, formatPrice, addToCart, isInCart };

  if (variant === 'Журнал') {
    return (
      <React.Fragment>
        <Hero {...commonProps} />
        <HomeNewSection {...commonProps} />
        <HomeTestimonialsSection testimonials={data.testimonials} />
        <HomeAuthorSection author={data.author} />
        <HomeGalleryStrip gallery={data.gallery} />
      </React.Fragment>
    );
  }

  if (variant === 'Витрина') {
    return (
      <React.Fragment>
        <Hero {...commonProps} />
        <HomeNewSection {...commonProps} />
        <HomeAuthorSection author={data.author} />
        <HomeTestimonialsSection testimonials={data.testimonials} />
      </React.Fragment>
    );
  }

  return (
    <React.Fragment>
      <Hero {...commonProps} />
      <HomeNewSection {...commonProps} />
      <HomeAuthorSection author={data.author} />
      <HomeTestimonialsSection testimonials={data.testimonials} />
      <HomeGalleryStrip gallery={data.gallery} />
    </React.Fragment>
  );
}
