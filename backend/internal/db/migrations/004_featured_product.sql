ALTER TABLE site_content
  ADD COLUMN featured_product_id text REFERENCES products(id) ON DELETE SET NULL;

UPDATE site_content
SET featured_product_id = 'lighthouse_aniva'
WHERE id = true AND featured_product_id IS NULL;
