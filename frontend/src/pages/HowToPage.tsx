import React from 'react';
import { useNavigate } from 'react-router-dom';
import { ProductCard, Stitch } from '../components/index';
import { productPath, ROUTES } from '../utils/routes';
import type { FormatPrice, HowToStep, ProductIdHandler, SiteData } from '../types/site';

interface StepCardProps {
  step: HowToStep;
}

function StepCard({ step }: StepCardProps) {
  return (
    <div className="step-card">
      <div className="step-n">{step.n}</div>
      <h3>{step.t}</h3>
      <p>{step.d}</p>
    </div>
  );
}

function QuestionCta() {
  return (
    <div className="sec-tint question-cta">
      <div>
        <h2 className="h-sec question-title">Остались вопросы?</h2>
        <p className="question-text">Напишите — Екатерина отвечает лично, обычно в течение дня.</p>
      </div>
      <a className="btn btn-primary" href="http://vk.com/id22478488" target="_blank" rel="noopener">Задать вопрос</a>
    </div>
  );
}

interface HowToPageProps {
  data: SiteData;
  formatPrice: FormatPrice;
  addToCart: ProductIdHandler;
  isInCart: (productId: string) => boolean;
}

export function HowToPage({ data, formatPrice, addToCart, isInCart }: HowToPageProps) {
  const navigate = useNavigate();
  const starters = data.products.filter((product) => product.price <= 200).slice(0, 4);

  return (
    <div data-screen-label="Как купить">
      <div className="wrap page-head">
        <Stitch />
        <h1 className="h-sec page-title">Как купить схему</h1>
        <p className="lede page-lede wide">
          Все схемы электронные — вам не нужно ждать доставку. Четыре простых шага, и схема у вас на почте.
        </p>
      </div>
      <div className="wrap page-content">
        <div className="steps">
          {data.howToBuy.map((step) => <StepCard key={step.n} step={step} />)}
        </div>
        <QuestionCta />
        <div className="starter-section">
          <div className="sec-head starter-head">
            <h2 className="h-sec starter-title">Начните с этих схем</h2>
            <button className="btn btn-ghost" onClick={() => navigate(ROUTES.shop)}>Весь каталог →</button>
          </div>
          <div className="pgrid">
            {starters.map((product) => (
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
      </div>
    </div>
  );
}
