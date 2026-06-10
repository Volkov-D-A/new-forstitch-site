package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"new-forstitch-site/backend/internal/models"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (s *PostgresRepository) Categories() []models.Category {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, label
		FROM categories
		ORDER BY sort_order, label
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		if err := rows.Scan(&category.ID, &category.Label); err != nil {
			return nil
		}
		categories = append(categories, category)
	}
	return categories
}

func (s *PostgresRepository) CreateCategory(category models.Category) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO categories (id, label)
		VALUES ($1, $2)
	`, category.ID, category.Label)
	return mapPostgresError(err)
}

func (s *PostgresRepository) UpdateCategory(id string, category models.Category) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE categories
		SET label = $2
		WHERE id = $1
	`, id, category.Label)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "category_not_found", "category not found")
}

func (s *PostgresRepository) DeleteCategory(id string) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM categories
		WHERE id = $1
	`, id)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "category_not_found", "category not found")
}

func (s *PostgresRepository) Products() []models.Product {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id,
		       title,
		       price,
		       cat,
		       sub,
		       img,
		       id IN (
		         SELECT id
		         FROM products
		         WHERE published = true
		         ORDER BY created_at DESC, id DESC
		         LIMIT 4
		       ) AS is_new,
		       size,
		       colors,
		       canvas
		FROM products
		WHERE published = true
		ORDER BY sort_order, title
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		product, err := scanProduct(rows)
		if err != nil {
			return nil
		}
		products = append(products, product)
	}
	return products
}

func (s *PostgresRepository) Product(id string) (models.Product, error) {
	row := s.db.QueryRowContext(context.Background(), `
		SELECT id,
		       title,
		       price,
		       cat,
		       sub,
		       img,
		       id IN (
		         SELECT id
		         FROM products
		         WHERE published = true
		         ORDER BY created_at DESC, id DESC
		         LIMIT 4
		       ) AS is_new,
		       size,
		       colors,
		       canvas
		FROM products
		WHERE id = $1 AND published = true
	`, id)

	product, err := scanProduct(row)
	if err == sql.ErrNoRows {
		return models.Product{}, models.NotFound("product_not_found", "product not found")
	}
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (s *PostgresRepository) CreateProduct(product models.Product) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO products (id, title, price, cat, sub, img, size, colors, canvas)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, product.ID, product.Title, product.Price, product.Cat, product.Sub, product.Img, product.Size, product.Colors, product.Canvas)
	return mapPostgresError(err)
}

func (s *PostgresRepository) UpdateProduct(id string, product models.Product) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE products
		SET title = $2,
		    price = $3,
		    cat = $4,
		    sub = $5,
		    img = $6,
		    size = $7,
		    colors = $8,
		    canvas = $9,
		    updated_at = now()
		WHERE id = $1
	`, id, product.Title, product.Price, product.Cat, product.Sub, product.Img, product.Size, product.Colors, product.Canvas)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "product_not_found", "product not found")
}

func (s *PostgresRepository) UpdateProductImage(id string, imageURL string) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE products
		SET img = $2,
		    updated_at = now()
		WHERE id = $1
	`, id, imageURL)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "product_not_found", "product not found")
}

func (s *PostgresRepository) DeleteProduct(id string) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM products
		WHERE id = $1
	`, id)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "product_not_found", "product not found")
}

func (s *PostgresRepository) Gallery() []models.GalleryItem {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, img, title, by_name
		FROM gallery_items
		WHERE published = true
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var gallery []models.GalleryItem
	for rows.Next() {
		var item models.GalleryItem
		if err := rows.Scan(&item.ID, &item.Img, &item.Title, &item.By); err != nil {
			return nil
		}
		gallery = append(gallery, item)
	}
	return gallery
}

func (s *PostgresRepository) CreateGalleryItem(item models.GalleryItem) (models.GalleryItem, error) {
	err := s.db.QueryRowContext(context.Background(), `
		INSERT INTO gallery_items (img, title, by_name)
		VALUES ($1, $2, $3)
		RETURNING id
	`, item.Img, item.Title, item.By).Scan(&item.ID)
	if err != nil {
		return models.GalleryItem{}, mapPostgresError(err)
	}
	return item, nil
}

func (s *PostgresRepository) UpdateGalleryItem(id int64, item models.GalleryItem) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE gallery_items
		SET img = $2,
		    title = $3,
		    by_name = $4
		WHERE id = $1
	`, id, item.Img, item.Title, item.By)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "gallery_item_not_found", "gallery item not found")
}

func (s *PostgresRepository) UpdateGalleryItemImage(id int64, imageURL string) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE gallery_items
		SET img = $2
		WHERE id = $1
	`, id, imageURL)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "gallery_item_not_found", "gallery item not found")
}

func (s *PostgresRepository) DeleteGalleryItem(id int64) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM gallery_items
		WHERE id = $1
	`, id)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "gallery_item_not_found", "gallery item not found")
}

func (s *PostgresRepository) Blog() []models.BlogPost {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, title, post_date, tag, img, excerpt, content
		FROM blog_posts
		WHERE published = true
		ORDER BY post_date DESC, id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var posts []models.BlogPost
	for rows.Next() {
		var post models.BlogPost
		var postDate time.Time
		if err := rows.Scan(&post.ID, &post.Title, &postDate, &post.Tag, &post.Img, &post.Excerpt, &post.Content); err != nil {
			return nil
		}
		post.Date = postDate.Format("2006-01-02")
		posts = append(posts, post)
	}
	return posts
}

func (s *PostgresRepository) CreateBlogPost(post models.BlogPost) (models.BlogPost, error) {
	postDate, err := parsePostDate(post.Date)
	if err != nil {
		return models.BlogPost{}, err
	}
	err = s.db.QueryRowContext(context.Background(), `
		INSERT INTO blog_posts (id, title, post_date, tag, img, excerpt, content)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING post_date
	`, post.ID, post.Title, postDate, post.Tag, post.Img, post.Excerpt, post.Content).Scan(&postDate)
	if err != nil {
		return models.BlogPost{}, mapPostgresError(err)
	}
	post.Date = postDate.Format("2006-01-02")
	return post, nil
}

func (s *PostgresRepository) UpdateBlogPost(id string, post models.BlogPost) error {
	postDate, err := parsePostDate(post.Date)
	if err != nil {
		return err
	}
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE blog_posts
		SET title = $2,
		    post_date = $3,
		    tag = $4,
		    img = $5,
		    excerpt = $6,
		    content = $7
		WHERE id = $1
	`, id, post.Title, postDate, post.Tag, post.Img, post.Excerpt, post.Content)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "blog_post_not_found", "blog post not found")
}

func (s *PostgresRepository) UpdateBlogPostImage(id string, imageURL string) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE blog_posts
		SET img = $2
		WHERE id = $1
	`, id, imageURL)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "blog_post_not_found", "blog post not found")
}

func (s *PostgresRepository) DeleteBlogPost(id string) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM blog_posts
		WHERE id = $1
	`, id)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "blog_post_not_found", "blog post not found")
}

func (s *PostgresRepository) SiteContent() models.SiteContent {
	var content models.SiteContent
	var featuredProductID sql.NullString
	err := s.db.QueryRowContext(context.Background(), `
		SELECT author_name, author_photo, author_p1, author_p2, author_p3, author_sign, featured_product_id
		FROM site_content
		WHERE id = true
	`).Scan(
		&content.Author.Name,
		&content.Author.Photo,
		&content.Author.P1,
		&content.Author.P2,
		&content.Author.P3,
		&content.Author.Sign,
		&featuredProductID,
	)
	if err != nil {
		return content
	}
	if featuredProductID.Valid {
		content.FeaturedProductID = featuredProductID.String
	}

	content.HowToBuy = s.howToSteps()
	content.Testimonials = s.Testimonials()
	return content
}

func (s *PostgresRepository) UpdateSiteSettings(settings models.SiteSettings) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE site_content
		SET featured_product_id = NULLIF($1, '')
		WHERE id = true
	`, settings.FeaturedProductID)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "site_content_not_found", "site content not found")
}

func (s *PostgresRepository) CreateOrder(req models.OrderRequest) models.OrderResponse {
	ctx := context.Background()
	orderID := fmt.Sprintf("order_%d", time.Now().UnixNano())
	message := "Заказ создан. Оплата будет подключена позже."

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.OrderResponse{ID: orderID, Message: "Не удалось создать заказ."}
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO orders (id, message) VALUES ($1, $2)
	`, orderID, message); err != nil {
		return models.OrderResponse{ID: orderID, Message: "Не удалось создать заказ."}
	}

	for _, item := range req.Items {
		var price int
		if err := tx.QueryRowContext(ctx, `
			SELECT price FROM products WHERE id = $1 AND published = true
		`, item.ProductID).Scan(&price); err != nil {
			return models.OrderResponse{ID: orderID, Message: "Не удалось создать заказ."}
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, price); err != nil {
			return models.OrderResponse{ID: orderID, Message: "Не удалось создать заказ."}
		}
	}

	if err := tx.Commit(); err != nil {
		return models.OrderResponse{ID: orderID, Message: "Не удалось создать заказ."}
	}

	return models.OrderResponse{ID: orderID, Message: message}
}

func (s *PostgresRepository) AdminUserByUsername(username string) (models.AdminUser, error) {
	var user models.AdminUser
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, username, password_hash
		FROM admin_users
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err == sql.ErrNoRows {
		return models.AdminUser{}, models.NotFound("admin_user_not_found", "admin user not found")
	}
	return user, err
}

func (s *PostgresRepository) EnsureAdminUser(username string, passwordHash string) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO admin_users (username, password_hash)
		VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
		    updated_at = now()
	`, username, passwordHash)
	return mapPostgresError(err)
}

func (s *PostgresRepository) CreateAdminSession(session models.AdminSession, expiresAt time.Time) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO admin_sessions (id, user_id, csrf_token, expires_at)
		VALUES ($1, $2, $3, $4)
	`, session.ID, session.UserID, session.CSRFToken, expiresAt)
	return mapPostgresError(err)
}

func (s *PostgresRepository) AdminSession(sessionID string, now time.Time) (models.AdminSession, error) {
	var session models.AdminSession
	err := s.db.QueryRowContext(context.Background(), `
		SELECT s.id, s.user_id, u.username, s.csrf_token
		FROM admin_sessions s
		JOIN admin_users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.expires_at > $2
	`, sessionID, now).Scan(&session.ID, &session.UserID, &session.Username, &session.CSRFToken)
	if err == sql.ErrNoRows {
		return models.AdminSession{}, models.Unauthorized("session_invalid", "admin session is invalid")
	}
	return session, err
}

func (s *PostgresRepository) DeleteAdminSession(sessionID string) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM admin_sessions
		WHERE id = $1
	`, sessionID)
	return err
}

func (s *PostgresRepository) DeleteExpiredAdminSessions(now time.Time) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM admin_sessions
		WHERE expires_at <= $1
	`, now)
	return err
}

type productScanner interface {
	Scan(dest ...any) error
}

func scanProduct(scanner productScanner) (models.Product, error) {
	var product models.Product
	err := scanner.Scan(
		&product.ID,
		&product.Title,
		&product.Price,
		&product.Cat,
		&product.Sub,
		&product.Img,
		&product.IsNew,
		&product.Size,
		&product.Colors,
		&product.Canvas,
	)
	return product, err
}

func (s *PostgresRepository) howToSteps() []models.HowToStep {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT n, title, description
		FROM how_to_steps
		ORDER BY sort_order, n
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var steps []models.HowToStep
	for rows.Next() {
		var step models.HowToStep
		if err := rows.Scan(&step.N, &step.T, &step.D); err != nil {
			return nil
		}
		steps = append(steps, step)
	}
	return steps
}

func (s *PostgresRepository) Testimonials() []models.Testimonial {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, name, role, img, text
		FROM testimonials
		WHERE published = true
		ORDER BY sort_order, id
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var testimonials []models.Testimonial
	for rows.Next() {
		var testimonial models.Testimonial
		if err := rows.Scan(&testimonial.ID, &testimonial.Name, &testimonial.Role, &testimonial.Img, &testimonial.Text); err != nil {
			return nil
		}
		testimonials = append(testimonials, testimonial)
	}
	return testimonials
}

func (s *PostgresRepository) CreateTestimonial(testimonial models.Testimonial) (models.Testimonial, error) {
	err := s.db.QueryRowContext(context.Background(), `
		INSERT INTO testimonials (name, role, img, text)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, testimonial.Name, testimonial.Role, testimonial.Img, testimonial.Text).Scan(&testimonial.ID)
	if err != nil {
		return models.Testimonial{}, mapPostgresError(err)
	}
	return testimonial, nil
}

func (s *PostgresRepository) UpdateTestimonial(id int64, testimonial models.Testimonial) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE testimonials
		SET name = $2,
		    role = $3,
		    img = $4,
		    text = $5
		WHERE id = $1
	`, id, testimonial.Name, testimonial.Role, testimonial.Img, testimonial.Text)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "testimonial_not_found", "testimonial not found")
}

func (s *PostgresRepository) UpdateTestimonialImage(id int64, imageURL string) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE testimonials
		SET img = $2
		WHERE id = $1
	`, id, imageURL)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "testimonial_not_found", "testimonial not found")
}

func (s *PostgresRepository) DeleteTestimonial(id int64) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM testimonials
		WHERE id = $1
	`, id)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "testimonial_not_found", "testimonial not found")
}

func parsePostDate(value string) (time.Time, error) {
	postDate, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, models.Validation("date_invalid", "date must use YYYY-MM-DD format")
	}
	return postDate, nil
}

func requireAffected(result sql.Result, code string, message string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return models.NotFound(code, message)
	}
	return nil
}

func mapPostgresError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case "23503":
		return models.Validation("reference_not_found", "referenced record not found")
	case "23505":
		return models.Conflict("record_exists", "record already exists")
	}

	return err
}
