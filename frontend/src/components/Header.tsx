import React from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { categoryPath, ROUTES } from '../utils/routes';
import { Logo } from './Logo';
import type { Category, Product } from '../types/site';

const navLinks = [
  { to: ROUTES.home, label: 'Главная', end: true },
  { to: ROUTES.gallery, label: 'Галерея' },
  { to: ROUTES.blog, label: 'Блог' },
  { to: ROUTES.howToBuy, label: 'Как купить' },
];

interface HeaderProps {
  cartCount: number;
  onCart: () => void;
  categories: Category[];
  products: Product[];
}

export function Header({ cartCount, onCart, categories, products }: HeaderProps) {
  const [shopOpen, setShopOpen] = React.useState(false);
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const dropdownRef = React.useRef<HTMLDivElement | null>(null);
  const location = useLocation();

  React.useEffect(() => {
    const close = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node | null)) setShopOpen(false);
    };
    document.addEventListener('click', close);
    return () => document.removeEventListener('click', close);
  }, []);

  const counts = products.reduce<Record<string, number>>((acc, product) => {
    acc[product.cat] = (acc[product.cat] || 0) + 1;
    return acc;
  }, {});
  const isShopActive = location.pathname.startsWith('/shop') || location.pathname.startsWith('/product');
  const closeMenus = () => {
    setShopOpen(false);
    setMobileOpen(false);
  };

  return (
    <header className="hdr">
      <div className="wrap hdr-in">
        <Logo onClick={closeMenus} />
        <nav className="nav">
          <div className="dropdown" ref={dropdownRef}>
            <button
              className={'nav-link' + (isShopActive ? ' active' : '')}
              onClick={() => setShopOpen(!shopOpen)}
            >
              Магазин ▾
            </button>
            {shopOpen ? (
              <div className="dropdown-menu">
                {categories.map((category) => (
                  <NavLink
                    key={category.id}
                    className="dropdown-item"
                    to={categoryPath(category.id)}
                    onClick={closeMenus}
                  >
                    <span>{category.label}</span>
                    <span className="cnt">
                      {category.id === 'all' ? products.length : (counts[category.id] || 0)}
                    </span>
                  </NavLink>
                ))}
              </div>
            ) : null}
          </div>
          {navLinks.map((link) => (
            <NavLink
              key={link.to}
              className={({ isActive }) => 'nav-link' + (isActive ? ' active' : '')}
              to={link.to}
              end={link.end}
              onClick={closeMenus}
            >
              {link.label}
            </NavLink>
          ))}
        </nav>
        <div className="hdr-actions">
          <button className="icon-btn" title="Корзина" onClick={onCart}>
            <svg width="21" height="21" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"></circle><circle cx="20" cy="21" r="1"></circle><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"></path></svg>
            {cartCount > 0 ? <span className="cart-badge">{cartCount}</span> : null}
          </button>
          <button className="icon-btn burger" title="Меню" onClick={() => setMobileOpen(true)}>
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="3" y1="7" x2="21" y2="7"></line><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="17" x2="21" y2="17"></line></svg>
          </button>
        </div>
      </div>
      {mobileOpen ? (
        <div className="mobile-menu">
          <div className="mobile-menu-head">
            <Logo onClick={closeMenus} />
            <button className="icon-btn" onClick={() => setMobileOpen(false)}>
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><line x1="5" y1="5" x2="19" y2="19"></line><line x1="19" y1="5" x2="5" y2="19"></line></svg>
            </button>
          </div>
          <NavLink className="mobile-link" to={ROUTES.home} onClick={closeMenus} end>Главная</NavLink>
          <NavLink className="mobile-link" to={ROUTES.shop} onClick={closeMenus}>Магазин</NavLink>
          <NavLink className="mobile-link" to={ROUTES.gallery} onClick={closeMenus}>Галерея</NavLink>
          <NavLink className="mobile-link" to={ROUTES.blog} onClick={closeMenus}>Блог</NavLink>
          <NavLink className="mobile-link" to={ROUTES.howToBuy} onClick={closeMenus}>Как купить</NavLink>
          <div className="mobile-category-list">
            {categories.filter((category) => category.id !== 'all').map((category) => (
              <NavLink key={category.id} className="chip" to={categoryPath(category.id)} onClick={closeMenus}>
                {category.label}
              </NavLink>
            ))}
          </div>
        </div>
      ) : null}
    </header>
  );
}
