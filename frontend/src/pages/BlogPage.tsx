import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { EmptyState, SImg, Stitch } from '../components/index';
import { blogPostPath, ROUTES } from '../utils/routes';
import type { BlogPost, SiteData } from '../types/site';

interface BlogCardProps {
  post: BlogPost;
}

function shortExcerpt(text: string, maxLength = 180) {
  if (text.length <= maxLength) return text;
  const trimmed = text.slice(0, maxLength).trimEnd();
  const lastSpace = trimmed.lastIndexOf(' ');
  return `${trimmed.slice(0, lastSpace > 80 ? lastSpace : trimmed.length)}...`;
}

function BlogCard({ post }: BlogCardProps) {
  const navigate = useNavigate();
  const openPost = () => navigate(blogPostPath(post.id));

  return (
    <article className="blog-card" onClick={openPost} onKeyDown={(event) => {
      if (event.key === 'Enter' || event.key === ' ') openPost();
    }} role="link" tabIndex={0}>
      <SImg src={post.img} alt={post.title} loading="lazy" />
      <div className="blog-body">
        <span className="blog-tag">{post.tag}</span>
        <h2 className="blog-title">{post.title}</h2>
        <p className="blog-excerpt">{shortExcerpt(post.excerpt)}</p>
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

export function BlogPostPage({ data }: BlogPageProps) {
  const navigate = useNavigate();
  const { postId } = useParams();
  const post = data.blog.find((item) => item.id === postId);

  if (!post) {
    return (
      <EmptyState
        title="Запись не найдена"
        text="Возможно, запись была снята с публикации или ссылка устарела."
        action={<button className="btn btn-primary" onClick={() => navigate(ROUTES.blog)}>Вернуться в блог</button>}
      />
    );
  }

  const paragraphs = post.content
    .split(/\n+/)
    .map((paragraph) => paragraph.trim())
    .filter(Boolean);

  return (
    <article data-screen-label={'Блог: ' + post.title}>
      <div className="wrap page-head compact">
        <nav className="crumbs">
          <button onClick={() => navigate(ROUTES.home)}>Главная</button> /
          <button onClick={() => navigate(ROUTES.blog)}>Блог</button> /
          <span>{post.title}</span>
        </nav>
      </div>
      <div className="wrap blog-post">
        {post.img ? <SImg src={post.img} alt={post.title} /> : null}
        <div className="blog-post-body">
          <span className="blog-tag">{post.tag}</span>
          <h1 className="h-sec page-title">{post.title}</h1>
          <span className="blog-date">{post.date}</span>
          <div className="blog-post-content">
            {paragraphs.map((paragraph, index) => <p key={index}>{paragraph}</p>)}
          </div>
        </div>
      </div>
    </article>
  );
}
