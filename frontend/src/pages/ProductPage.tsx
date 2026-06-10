import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { EmptyState, ProductCard, SImg } from '../components/index';
import { categoryPath, productPath, ROUTES } from '../utils/routes';
import type { Category, FormatPrice, Product, ProductIdHandler, SiteData } from '../types/site';

interface ProductSpecsProps {
  product: Product;
  category?: Category;
}

function ProductSpecs({ product, category }: ProductSpecsProps) {
  return (
    <dl className="spec-table">
      <div className="spec-row"><dt>Размер</dt><dd>{product.size}</dd></div>
      <div className="spec-row"><dt>Палитра</dt><dd>{product.colors}</dd></div>
      <div className="spec-row"><dt>Рекомендуемая основа</dt><dd>{product.canvas}</dd></div>
      <div className="spec-row"><dt>Категория</dt><dd>{category ? category.label : ''}{product.sub ? ' · ' + product.sub : ''}</dd></div>
      <div className="spec-row"><dt>Формат</dt><dd>PDF: цветной ключ + символы</dd></div>
    </dl>
  );
}

interface RelatedProductsProps {
  products: Product[];
  categories: Category[];
  formatPrice: FormatPrice;
  addToCart: ProductIdHandler;
  isInCart: (productId: string) => boolean;
  onOpen: ProductIdHandler;
}

function RelatedProducts({ products, categories, formatPrice, addToCart, isInCart, onOpen }: RelatedProductsProps) {
  if (products.length === 0) return null;

  return (
    <section className="sec sec-tint related-section">
      <div className="wrap">
        <div className="sec-head related-head">
          <h2 className="h-sec related-title">Похожие схемы</h2>
        </div>
        <div className="pgrid">
          {products.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              categories={categories}
              formatPrice={formatPrice}
              onOpen={onOpen}
              onAdd={addToCart}
              inCart={isInCart(product.id)}
            />
          ))}
        </div>
      </div>
    </section>
  );
}

interface ProductPageProps {
  data: SiteData;
  formatPrice: FormatPrice;
  addToCart: ProductIdHandler;
  isInCart: (productId: string) => boolean;
}

export function ProductPage({ data, formatPrice, addToCart, isInCart }: ProductPageProps) {
  const navigate = useNavigate();
  const { productId } = useParams();
  const product = data.products.find((item) => item.id === productId);

  if (!product) {
    return (
      <EmptyState
        title="Схема не найдена"
        text="Возможно, товар был снят с публикации или ссылка устарела."
        action={<button className="btn btn-primary" onClick={() => navigate(ROUTES.shop)}>Вернуться в магазин</button>}
      />
    );
  }

  const category = data.categories.find((item) => item.id === product.cat);
  const relatedProducts = data.products
    .filter((item) => item.cat === product.cat && item.id !== product.id)
    .slice(0, 4);
  const inCart = isInCart(product.id);
  const openProduct = (id: string) => navigate(productPath(id));

  return (
    <div data-screen-label={'Товар: ' + product.title}>
      <div className="wrap page-head compact">
        <nav className="crumbs">
          <button onClick={() => navigate(ROUTES.home)}>Главная</button> /
          <button onClick={() => navigate(ROUTES.shop)}>Магазин</button> /
          <button onClick={() => navigate(categoryPath(product.cat))}>{category ? category.label : ''}</button> /
          <span>{product.title}</span>
        </nav>
      </div>
      <div className="wrap product-layout with-bottom-space">
        <div className="product-img">
          {product.img
            ? <SImg src={product.img} alt={product.title} />
            : <div className="ph-img product-placeholder"><span className="x-mark">× × ×</span><span>фото готовится</span></div>}
        </div>
        <div>
          {product.isNew ? <span className="pcard-badge product-badge-inline">Новинка</span> : null}
          <h1 className="h-sec product-title">{product.title}</h1>
          <p className="product-price">{formatPrice(product.price)}</p>
          <p className="product-subtitle">электронная схема · PDF · мгновенная доставка</p>
          <ProductSpecs product={product} category={category} />
          <div className="product-actions">
            <button
              className={'btn ' + (inCart ? 'btn-outline' : 'btn-primary')}
              onClick={() => addToCart(product.id)}
            >
              {inCart ? '✓ В корзине' : 'Добавить в корзину'}
            </button>
            <button className="btn btn-ghost" onClick={() => navigate(ROUTES.howToBuy)}>Как проходит покупка?</button>
          </div>
          <div className="x-row product-note">
            <span className="x-mark soft">× × ×</span>
            <span className="product-note-text">схема разработана вручную, проверена отшивом</span>
          </div>
        </div>
      </div>
      <RelatedProducts
        products={relatedProducts}
        categories={data.categories}
        formatPrice={formatPrice}
        addToCart={addToCart}
        isInCart={isInCart}
        onOpen={openProduct}
      />
    </div>
  );
}
