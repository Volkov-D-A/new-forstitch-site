import React from 'react';
import { SImg } from '../index';
import type { Author } from '../../types/site';

interface HomeAuthorSectionProps {
  author: Author;
}

export function HomeAuthorSection({ author }: HomeAuthorSectionProps) {
  return (
    <section className="sec sec-paper" data-screen-label="Главная: об авторе">
      <div className="wrap author-grid">
        <div className="author-photo">
          <SImg src={author.photo} alt={author.name} loading="lazy" />
        </div>
        <div className="author-text">
          <p className="eyebrow">Об авторе</p>
          <h2 className="h-sec">{author.name}</h2>
          <p className="author-first">{author.p1}</p>
          <p>{author.p2}</p>
          <p>{author.p3}</p>
          <p className="hand author-sign">{author.sign}</p>
        </div>
      </div>
    </section>
  );
}
