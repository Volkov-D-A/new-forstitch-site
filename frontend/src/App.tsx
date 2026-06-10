// Корневое приложение forstitch redesign
import React from 'react';
import { Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import { CartDrawer, Footer, Header, Toast } from './components/index';
import { useCart } from './hooks/useCart';
import { useSiteData } from './hooks/useSiteData';
import { BlogPage, GalleryPage, HomePage, HowToPage, ProductPage, ShopPage } from './pages/index';
import { formatPrice } from './utils/currency';
import { HOME_VARIANT } from './utils/homeContent';
import { ROUTES } from './utils/routes';

function App() {
  const { data, error, isLoading } = useSiteData();
  const navigate = useNavigate();
  const location = useLocation();
  const [toast, setToast] = React.useState<string | null>(null);
  const toastTimer = React.useRef<number | undefined>(undefined);
  const showToast = (text: string) => {
    setToast(text);
    clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(null), 2200);
  };
  const {
    addToCart,
    cart,
    closeCart,
    isCartOpen,
    openCart,
    removeFromCart,
  } = useCart({
    products: data?.products || [],
    onAdded: (product) => showToast('«' + (product ? product.title : '') + '» — в корзине'),
  });

  React.useEffect(() => {
    window.scrollTo({ top: 0 });
  }, [location.pathname]);

  if (isLoading) {
    return <div className="app-state">Загружаем каталог...</div>;
  }

  if (error) {
    return <div className="app-state">Не удалось загрузить данные сайта.</div>;
  }

  if (!data) {
    return <div className="app-state">Данные сайта не найдены.</div>;
  }

  return (
    <React.Fragment>
      <Header cartCount={cart.length} onCart={openCart} categories={data.categories} products={data.products} />
      <main>
        <Routes>
          <Route path={ROUTES.home} element={<HomePage variant={HOME_VARIANT} addToCart={addToCart} cart={cart} data={data} formatPrice={formatPrice} />} />
          <Route path={ROUTES.shop} element={<ShopPage addToCart={addToCart} cart={cart} data={data} formatPrice={formatPrice} />} />
          <Route path="/shop/:categoryId" element={<ShopPage addToCart={addToCart} cart={cart} data={data} formatPrice={formatPrice} />} />
          <Route path="/product/:productId" element={<ProductPage addToCart={addToCart} cart={cart} data={data} formatPrice={formatPrice} />} />
          <Route path={ROUTES.gallery} element={<GalleryPage data={data} />} />
          <Route path={ROUTES.blog} element={<BlogPage data={data} />} />
          <Route path={ROUTES.howToBuy} element={<HowToPage addToCart={addToCart} cart={cart} data={data} formatPrice={formatPrice} />} />
          <Route path="/howto" element={<Navigate to={ROUTES.howToBuy} replace />} />
          <Route path="*" element={<Navigate to={ROUTES.home} replace />} />
        </Routes>
      </main>
      <Footer categories={data.categories} />
      {isCartOpen ? (
        <CartDrawer
          cart={cart}
          onClose={closeCart}
          onRemove={removeFromCart}
          onShopOpen={() => {
            closeCart();
            navigate(ROUTES.shop);
          }}
          products={data.products}
          formatPrice={formatPrice}
        />
      ) : null}
      {toast ? <Toast text={toast} /> : null}
    </React.Fragment>
  );
}

export default App;
