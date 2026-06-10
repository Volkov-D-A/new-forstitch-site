ALTER TABLE blog_posts
  ADD COLUMN content text NOT NULL DEFAULT '';

UPDATE blog_posts
SET content = excerpt
WHERE content = '';
