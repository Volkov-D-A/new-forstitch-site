import React from 'react';
import { SImg, Stitch } from '../components/index';
import type { BlogPost, SiteData } from '../types/site';

interface BlogCardProps {
  post: BlogPost;
}

function BlogCard({ post }: BlogCardProps) {
  return (
    <article className="blog-card">
      <SImg src={post.img} alt={post.title} loading="lazy" />
      <div className="blog-body">
        <span className="blog-tag">{post.tag}</span>
        <h2 className="blog-title">{post.title}</h2>
        <p className="blog-excerpt">{post.excerpt}</p>
        <span className="blog-date">{post.date}</span>
      </div>
    </article>
  );
}

interface BlogPageProps {
  data: SiteData;
}

export function BlogPage({ data }: BlogPageProps) {
  return (
    <div data-screen-label="Блог">
      <div className="wrap page-head">
        <Stitch />
        <h1 className="h-sec page-title">Блог</h1>
        <p className="lede page-lede">
          Новинки, процессы разработки и советы по вышивке сложных авторских схем.
        </p>
      </div>
      <div className="wrap page-content">
        <div className="blog-grid">
          {data.blog.map((post) => <BlogCard key={post.id} post={post} />)}
        </div>
      </div>
    </div>
  );
}
