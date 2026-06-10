import React from 'react';
import { Link } from 'react-router-dom';
import { categoryPath, ROUTES } from '../utils/routes';
import type { Category } from '../types/site';

interface FooterProps {
  categories: Category[];
}

export function Footer({ categories }: FooterProps) {
  return (
    <footer className="ftr">
      <div className="wrap">
        <div className="ftr-grid">
          <div>
            <span className="logo-word">f<span className="logo-x ftr-logo-x">×</span>rstitch</span>
            <p className="ftr-about">
              Авторские схемы для вышивки крестом Екатерины Волковой. Живопись и фотография, перенесённые в крестики — с 2011 года.
            </p>
          </div>
          <div>
            <h4>Магазин</h4>
            {categories.filter((category) => category.id !== 'all').map((category) => (
              <Link key={category.id} className="ftr-link" to={categoryPath(category.id)}>{category.label}</Link>
            ))}
          </div>
          <div>
            <h4>Разделы</h4>
            <Link className="ftr-link" to={ROUTES.gallery}>Галерея отшивов</Link>
            <Link className="ftr-link" to={ROUTES.blog}>Блог</Link>
            <Link className="ftr-link" to={ROUTES.howToBuy}>Как купить</Link>
            <Link className="ftr-link" to={ROUTES.howToBuy}>Задать вопрос</Link>
          </div>
          <div>
            <h4>Связь</h4>
            <a className="ftr-link" href="http://vk.com/id22478488" target="_blank" rel="noopener">ВКонтакте</a>
            <span className="ftr-link ftr-static">forstitch.ru</span>
          </div>
        </div>
        <div className="ftr-bottom">
          <span>© Екатерина Волкова, 2011–2026 · Все права защищены</span>
          <span className="x-mark gold-soft">× × × × ×</span>
        </div>
      </div>
    </footer>
  );
}
