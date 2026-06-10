import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ProductCard, Stitch } from '../components/index';
import { plural } from '../utils/format';
import { categoryPath, productPath } from '../utils/routes';
import type { Category, CategoryId, FormatPrice, Product, ProductIdHandler, SiteData } from '../types/site';

type ProductSort = 'default' | 'new' | 'price-asc' | 'price-desc' | 'title';

function getSortedProducts(products: Product[], sort: ProductSort) {
  if (sort === 'price-asc') return [...products].sort((a, b) => a.price - b.price);
  if (sort === 'price-desc') return [...products].sort((a, b) => b.price - a.price);
  if (sort === 'title') return [...products].sort((a, b) => a.title.localeCompare(b.title, 'ru'));
  if (sort === 'new') return [...products].sort((a, b) => (b.isNew ? 1 : 0) - (a.isNew ? 1 : 0));
  return products;
}

interface CategoryFiltersProps {
  categories: Category[];
  products: Product[];
  activeCategoryId: CategoryId;
  onChange: (categoryId: CategoryId) => void;
}

function CategoryFilters({ categories, products, activeCategoryId, onChange }: CategoryFiltersProps) {
  return (
    <div className="filters">
      {categories.map((category) => {
        const count = category.id === 'all'
          ? products.length
          : products.filter((product) => product.cat === category.id).length;

        return (
          <button
            key={category.id}
            className={'chip' + (activeCategoryId === category.id ? ' on' : '')}
            onClick={() => onChange(category.id)}
          >
            {category.label} · {count}
          </button>
        );
      })}
    </div>
  );
}

interface SortSelectProps {
  value: ProductSort;
  onChange: (value: ProductSort) => void;
}

function SortSelect({ value, onChange }: SortSelectProps) {
  return (
    <select className="sort-sel" value={value} onChange={(event) => onChange(event.target.value as ProductSort)}>
      <option value="default">По умолчанию</option>
      <option value="new">Сначала новинки</option>
      <option value="price-asc">Цена: по возрастанию</option>
      <option value="price-desc">Цена: по убыванию</option>
      <option value="title">По названию</option>
    </select>
  );
}

interface ShopPageProps {
  data: SiteData;
  formatPrice: FormatPrice;
  addToCart: ProductIdHandler;
  cart: string[];
}

export function ShopPage({ data, formatPrice, addToCart, cart }: ShopPageProps) {
  const navigate = useNavigate();
  const { categoryId = 'all' } = useParams();
  const [sort, setSort] = React.useState<ProductSort>('default');
  const category = data.categories.find((item) => item.id === categoryId) || data.categories[0];
  const filteredProducts = data.products.filter((product) => category.id === 'all' || product.cat === category.id);
  const products = getSortedProducts(filteredProducts, sort);

  return (
    <div data-screen-label="Магазин">
      <div className="wrap page-head">
        <Stitch />
        <h1 className="h-sec page-title">
          {category.id === 'all' ? 'Магазин схем' : category.label}
        </h1>
        <p className="lede page-lede">
          Все схемы — электронные, в формате PDF: ключ, символьная схема и инструкция. Приходят на почту сразу после оплаты.
        </p>
      </div>
      <div className="wrap">
        <CategoryFilters
          categories={data.categories}
          products={data.products}
          activeCategoryId={category.id}
          onChange={(id: CategoryId) => navigate(categoryPath(id))}
        />
        <div className="shop-bar">
          <span className="muted-count">
            {products.length} {plural(products.length, 'схема', 'схемы', 'схем')}
          </span>
          <SortSelect value={sort} onChange={setSort} />
        </div>
        <div className="pgrid page-content">
          {products.map((product) => (
            <ProductCard
              key={product.id}
              product={product}
              categories={data.categories}
              formatPrice={formatPrice}
              onOpen={(id: string) => navigate(productPath(id))}
              onAdd={addToCart}
              inCart={cart.includes(product.id)}
            />
          ))}
        </div>
      </div>
    </div>
  );
}
