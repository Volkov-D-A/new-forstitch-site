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
		       description
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
		product.Images = s.productImages(product.ID)
		product.Files = s.productFiles(product.ID)
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
		       description
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
	product.Images = s.productImages(product.ID)
	product.Files = s.productFiles(product.ID)
	return product, nil
}

func (s *PostgresRepository) CreateProduct(product models.Product) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO products (id, title, price, cat, img, size, colors, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, product.ID, product.Title, product.Price, product.Cat, product.Img, product.Size, product.Colors, product.Description)
	return mapPostgresError(err)
}

func (s *PostgresRepository) UpdateProduct(id string, product models.Product) error {
	result, err := s.db.ExecContext(context.Background(), `
		UPDATE products
		SET title = $2,
		    price = $3,
		    cat = $4,
		    img = $5,
		    size = $6,
		    colors = $7,
		    description = $8,
		    updated_at = now()
		WHERE id = $1
	`, id, product.Title, product.Price, product.Cat, product.Img, product.Size, product.Colors, product.Description)
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

func (s *PostgresRepository) AddProductImage(productID string, imageURL string) (models.ProductImage, error) {
	var image models.ProductImage
	err := s.db.QueryRowContext(context.Background(), `
		INSERT INTO product_images (product_id, url, sort_order)
		VALUES ($1, $2, COALESCE((SELECT max(sort_order) + 10 FROM product_images WHERE product_id = $1), 10))
		RETURNING id, url
	`, productID, imageURL).Scan(&image.ID, &image.URL)
	return image, mapPostgresError(err)
}

func (s *PostgresRepository) DeleteProductImage(productID string, imageID int64) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM product_images
		WHERE product_id = $1 AND id = $2
	`, productID, imageID)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "product_image_not_found", "product image not found")
}

func (s *PostgresRepository) AddProductFile(productID string, name string, objectName string) (models.ProductFile, error) {
	var file models.ProductFile
	err := s.db.QueryRowContext(context.Background(), `
		INSERT INTO product_files (product_id, name, object_name, sort_order)
		VALUES ($1, $2, $3, COALESCE((SELECT max(sort_order) + 10 FROM product_files WHERE product_id = $1), 10))
		RETURNING id, name, object_name
	`, productID, name, objectName).Scan(&file.ID, &file.Name, &file.ObjectName)
	return file, mapPostgresError(err)
}

func (s *PostgresRepository) DeleteProductFile(productID string, fileID int64) error {
	result, err := s.db.ExecContext(context.Background(), `
		DELETE FROM product_files
		WHERE product_id = $1 AND id = $2
	`, productID, fileID)
	if err != nil {
		return mapPostgresError(err)
	}
	return requireAffected(result, "product_file_not_found", "product file not found")
}

func (s *PostgresRepository) ProductFileForCustomerOrder(orderID string, customerID int64, fileID int64) (models.ProductFile, error) {
	var file models.ProductFile
	err := s.db.QueryRowContext(context.Background(), `
		SELECT pf.id, pf.name, pf.object_name
		FROM product_files pf
		JOIN order_items oi ON oi.product_id = pf.product_id
		JOIN orders o ON o.id = oi.order_id
		WHERE o.id = $1
		  AND o.customer_id = $2
		  AND o.status IN ('paid', 'fulfilled')
		  AND pf.id = $3
	`, orderID, customerID, fileID).Scan(&file.ID, &file.Name, &file.ObjectName)
	if err == sql.ErrNoRows {
		return models.ProductFile{}, models.NotFound("product_file_not_found", "product file not found")
	}
	return file, err
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
		SELECT id, img, title, description
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
		if err := rows.Scan(&item.ID, &item.Img, &item.Title, &item.Description); err != nil {
			return nil
		}
		gallery = append(gallery, item)
	}
	return gallery
}

func (s *PostgresRepository) CreateGalleryItem(item models.GalleryItem) (models.GalleryItem, error) {
	err := s.db.QueryRowContext(context.Background(), `
		INSERT INTO gallery_items (img, title, description)
		VALUES ($1, $2, $3)
		RETURNING id
	`, item.Img, item.Title, item.Description).Scan(&item.ID)
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
		    description = $4
		WHERE id = $1
	`, id, item.Img, item.Title, item.Description)
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

func (s *PostgresRepository) CreateOrder(req models.OrderRequest, customer models.CustomerUser) (models.OrderResponse, error) {
	ctx := context.Background()
	orderID := fmt.Sprintf("order_%d", time.Now().UnixNano())
	message := "Заказ оформлен и считается оплаченным."

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.OrderResponse{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO orders (
			id, status, message, customer_id, customer_email, customer_name
		)
		VALUES ($1, 'paid', $2, $3, $4, $5)
	`, orderID, message, customer.ID, customer.Email, customer.Name); err != nil {
		return models.OrderResponse{}, mapPostgresError(err)
	}

	for _, item := range req.Items {
		var price int
		if err := tx.QueryRowContext(ctx, `
			SELECT price FROM products WHERE id = $1 AND published = true
		`, item.ProductID).Scan(&price); err != nil {
			if err == sql.ErrNoRows {
				return models.OrderResponse{}, models.NotFound("product_not_found", "product not found")
			}
			return models.OrderResponse{}, err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, price); err != nil {
			return models.OrderResponse{}, mapPostgresError(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return models.OrderResponse{}, err
	}

	return models.OrderResponse{ID: orderID, Status: "paid", Message: message}, nil
}

func (s *PostgresRepository) Orders() ([]models.Order, error) {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, status, customer_email, customer_name, message, created_at
		FROM orders
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items, err := s.orderItems(order.ID, order.Status)
		if err != nil {
			return nil, err
		}
		order.Items = items
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (s *PostgresRepository) CustomerOrders(customerID int64) ([]models.Order, error) {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, status, customer_email, customer_name, message, created_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		items, err := s.orderItems(order.ID, order.Status)
		if err != nil {
			return nil, err
		}
		order.Items = items
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (s *PostgresRepository) OrderForCustomer(orderID string, customerID int64) (models.Order, error) {
	return s.orderByQuery(`
		SELECT id, status, customer_email, customer_name, message, created_at
		FROM orders
		WHERE id = $1 AND customer_id = $2
	`, orderID, customerID)
}

func (s *PostgresRepository) orderByQuery(query string, args ...any) (models.Order, error) {
	row := s.db.QueryRowContext(context.Background(), query, args...)
	order, err := scanOrder(row)
	if err == sql.ErrNoRows {
		return models.Order{}, models.NotFound("order_not_found", "order not found")
	}
	if err != nil {
		return models.Order{}, err
	}
	items, err := s.orderItems(order.ID, order.Status)
	if err != nil {
		return models.Order{}, err
	}
	order.Items = items
	return order, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row rowScanner) (models.Order, error) {
	var order models.Order
	var createdAt time.Time
	if err := row.Scan(&order.ID, &order.Status, &order.CustomerEmail, &order.CustomerName, &order.Message, &createdAt); err != nil {
		return models.Order{}, err
	}
	order.CreatedAt = createdAt.Format(time.RFC3339)
	return order, nil
}

func (s *PostgresRepository) orderItems(orderID string, status string) ([]models.OrderItem, error) {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT oi.product_id, p.title, oi.quantity, oi.price
		FROM order_items oi
		JOIN products p ON p.id = oi.product_id
		WHERE oi.order_id = $1
		ORDER BY p.title
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Quantity, &item.Price); err != nil {
			return nil, err
		}
		if status == "paid" || status == "fulfilled" {
			for _, file := range s.productFiles(item.ProductID) {
				item.DownloadURLs = append(item.DownloadURLs, models.DownloadFile{
					ID:   file.ID,
					Name: file.Name,
					URL:  fmt.Sprintf("/api/customer/orders/%s/files/%d", orderID, file.ID),
				})
			}
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

func (s *PostgresRepository) CustomerByEmail(email string) (models.CustomerUser, error) {
	var user models.CustomerUser
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id, email, name, password_hash
		FROM customer_users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash)
	if err == sql.ErrNoRows {
		return models.CustomerUser{}, models.NotFound("customer_not_found", "customer not found")
	}
	return user, err
}

func (s *PostgresRepository) EnsureCustomer(email string, name string, passwordHash string) (models.CustomerUser, bool, error) {
	existing, err := s.CustomerByEmail(email)
	if err == nil {
		if name != "" && name != existing.Name {
			_, updateErr := s.db.ExecContext(context.Background(), `
				UPDATE customer_users
				SET name = $2, updated_at = now()
				WHERE id = $1
			`, existing.ID, name)
			if updateErr != nil {
				return models.CustomerUser{}, false, mapPostgresError(updateErr)
			}
			existing.Name = name
		}
		return existing, false, nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return models.CustomerUser{}, false, err
	}

	var user models.CustomerUser
	err = s.db.QueryRowContext(context.Background(), `
		INSERT INTO customer_users (email, name, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, email, name, password_hash
	`, email, name, passwordHash).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash)
	if err != nil {
		return models.CustomerUser{}, false, mapPostgresError(err)
	}
	return user, true, nil
}

func (s *PostgresRepository) SaveCustomerRegistrationCode(email string, name string, passwordHash string, codeHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO customer_registration_codes (email, name, password_hash, code_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (email) DO UPDATE
		SET name = EXCLUDED.name,
		    password_hash = EXCLUDED.password_hash,
		    code_hash = EXCLUDED.code_hash,
		    expires_at = EXCLUDED.expires_at,
		    created_at = now()
	`, email, name, passwordHash, codeHash, expiresAt)
	return mapPostgresError(err)
}

func (s *PostgresRepository) CustomerByRegistrationCode(email string, codeHash string, now time.Time) (models.CustomerUser, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.CustomerUser{}, err
	}
	defer tx.Rollback()

	var name string
	var passwordHash string
	err = tx.QueryRowContext(ctx, `
		SELECT name, password_hash
		FROM customer_registration_codes
		WHERE email = $1 AND code_hash = $2 AND expires_at > $3
	`, email, codeHash, now).Scan(&name, &passwordHash)
	if err == sql.ErrNoRows {
		return models.CustomerUser{}, models.NotFound("registration_code_not_found", "registration code not found")
	}
	if err != nil {
		return models.CustomerUser{}, err
	}

	var user models.CustomerUser
	err = tx.QueryRowContext(ctx, `
		INSERT INTO customer_users (email, name, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
		SET name = CASE WHEN EXCLUDED.name <> '' THEN EXCLUDED.name ELSE customer_users.name END,
		    updated_at = now()
		RETURNING id, email, name, password_hash
	`, email, name, passwordHash).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash)
	if err != nil {
		return models.CustomerUser{}, mapPostgresError(err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM customer_registration_codes
		WHERE email = $1
	`, email); err != nil {
		return models.CustomerUser{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.CustomerUser{}, err
	}
	return user, nil
}

func (s *PostgresRepository) DeleteCustomerRegistrationCode(email string) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM customer_registration_codes
		WHERE email = $1
	`, email)
	return err
}

func (s *PostgresRepository) SaveCustomerPasswordResetCode(email string, codeHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO customer_password_reset_codes (email, code_hash, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE
		SET code_hash = EXCLUDED.code_hash,
		    expires_at = EXCLUDED.expires_at,
		    created_at = now()
	`, email, codeHash, expiresAt)
	return mapPostgresError(err)
}

func (s *PostgresRepository) UpdateCustomerPasswordByResetCode(email string, codeHash string, passwordHash string, now time.Time) (models.CustomerUser, error) {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.CustomerUser{}, err
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT true
		FROM customer_password_reset_codes
		WHERE email = $1 AND code_hash = $2 AND expires_at > $3
	`, email, codeHash, now).Scan(&exists)
	if err == sql.ErrNoRows {
		return models.CustomerUser{}, models.NotFound("password_reset_code_not_found", "password reset code not found")
	}
	if err != nil {
		return models.CustomerUser{}, err
	}

	var user models.CustomerUser
	err = tx.QueryRowContext(ctx, `
		UPDATE customer_users
		SET password_hash = $2, updated_at = now()
		WHERE email = $1
		RETURNING id, email, name, password_hash
	`, email, passwordHash).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash)
	if err == sql.ErrNoRows {
		return models.CustomerUser{}, models.NotFound("customer_not_found", "customer not found")
	}
	if err != nil {
		return models.CustomerUser{}, mapPostgresError(err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM customer_password_reset_codes
		WHERE email = $1
	`, email); err != nil {
		return models.CustomerUser{}, err
	}
	if err := tx.Commit(); err != nil {
		return models.CustomerUser{}, err
	}
	return user, nil
}

func (s *PostgresRepository) DeleteCustomerPasswordResetCode(email string) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM customer_password_reset_codes
		WHERE email = $1
	`, email)
	return err
}

func (s *PostgresRepository) CreateCustomerSession(session models.CustomerSession, expiresAt time.Time) error {
	_, err := s.db.ExecContext(context.Background(), `
		INSERT INTO customer_sessions (id, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, session.ID, session.UserID, expiresAt)
	return mapPostgresError(err)
}

func (s *PostgresRepository) CustomerSession(sessionID string, now time.Time) (models.CustomerSession, error) {
	var session models.CustomerSession
	err := s.db.QueryRowContext(context.Background(), `
		SELECT s.id, s.user_id, u.email, u.name
		FROM customer_sessions s
		JOIN customer_users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.expires_at > $2
	`, sessionID, now).Scan(&session.ID, &session.UserID, &session.Email, &session.Name)
	if err == sql.ErrNoRows {
		return models.CustomerSession{}, models.Unauthorized("session_invalid", "customer session is invalid")
	}
	return session, err
}

func (s *PostgresRepository) DeleteCustomerSession(sessionID string) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM customer_sessions
		WHERE id = $1
	`, sessionID)
	return err
}

func (s *PostgresRepository) DeleteExpiredCustomerSessions(now time.Time) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM customer_sessions
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
		&product.Img,
		&product.IsNew,
		&product.Size,
		&product.Colors,
		&product.Description,
	)
	return product, err
}

func (s *PostgresRepository) productImages(productID string) []models.ProductImage {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, url
		FROM product_images
		WHERE product_id = $1
		ORDER BY sort_order, id
	`, productID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var images []models.ProductImage
	for rows.Next() {
		var image models.ProductImage
		if err := rows.Scan(&image.ID, &image.URL); err != nil {
			return nil
		}
		images = append(images, image)
	}
	return images
}

func (s *PostgresRepository) productFiles(productID string) []models.ProductFile {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id, name, object_name
		FROM product_files
		WHERE product_id = $1
		ORDER BY sort_order, id
	`, productID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var files []models.ProductFile
	for rows.Next() {
		var file models.ProductFile
		if err := rows.Scan(&file.ID, &file.Name, &file.ObjectName); err != nil {
			return nil
		}
		files = append(files, file)
	}
	return files
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
