package models

type Category struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Product struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Price       int            `json:"price"`
	Cat         string         `json:"cat"`
	Img         string         `json:"img,omitempty"`
	Images      []ProductImage `json:"images,omitempty"`
	IsNew       bool           `json:"isNew,omitempty"`
	Size        string         `json:"size"`
	Colors      string         `json:"colors"`
	Description string         `json:"description,omitempty"`
}

type ProductImage struct {
	ID  int64  `json:"id"`
	URL string `json:"url"`
}

type GalleryItem struct {
	ID    int64  `json:"id,omitempty"`
	Img   string `json:"img"`
	Title string `json:"title"`
	By    string `json:"by"`
}

type BlogPost struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Date    string `json:"date"`
	Tag     string `json:"tag"`
	Img     string `json:"img"`
	Excerpt string `json:"excerpt"`
	Content string `json:"content"`
}

type Author struct {
	Name  string `json:"name"`
	Photo string `json:"photo"`
	P1    string `json:"p1"`
	P2    string `json:"p2"`
	P3    string `json:"p3"`
	Sign  string `json:"sign"`
}

type HowToStep struct {
	N string `json:"n"`
	T string `json:"t"`
	D string `json:"d"`
}

type Testimonial struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
	Role string `json:"role"`
	Img  string `json:"img"`
	Text string `json:"text"`
}

type SiteContent struct {
	Author            Author        `json:"author"`
	FeaturedProductID string        `json:"featuredProductId,omitempty"`
	HowToBuy          []HowToStep   `json:"howToBuy"`
	Testimonials      []Testimonial `json:"testimonials"`
}

type SiteSettings struct {
	FeaturedProductID string `json:"featuredProductId"`
}

type CartItem struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type OrderRequest struct {
	Items []CartItem `json:"items"`
}

type OrderResponse struct {
	ID          string `json:"id"`
	CheckoutURL string `json:"checkoutUrl,omitempty"`
	Message     string `json:"message,omitempty"`
	Status      string `json:"status,omitempty"`
}

type Order struct {
	ID            string      `json:"id"`
	Status        string      `json:"status"`
	CustomerEmail string      `json:"customerEmail"`
	CustomerName  string      `json:"customerName,omitempty"`
	Message       string      `json:"message,omitempty"`
	Items         []OrderItem `json:"items"`
	CreatedAt     string      `json:"createdAt"`
}

type OrderItem struct {
	ProductID   string `json:"productId"`
	ProductName string `json:"productName"`
	Quantity    int    `json:"quantity"`
	Price       int    `json:"price"`
	DownloadURL string `json:"downloadUrl,omitempty"`
}

type CustomerUser struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
}

type CustomerSession struct {
	ID     string
	UserID int64
	Email  string
	Name   string
}

type CustomerSessionResponse struct {
	Authenticated bool   `json:"authenticated"`
	Email         string `json:"email,omitempty"`
	Name          string `json:"name,omitempty"`
}

type CustomerRegistrationStartRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Password string `json:"password"`
}

type CustomerRegistrationVerifyRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type CustomerRegistrationStartResponse struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

type CustomerPasswordResetStartRequest struct {
	Email string `json:"email"`
}

type CustomerPasswordResetVerifyRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
}

type AdminSession struct {
	ID        string
	UserID    int64
	Username  string
	CSRFToken string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Username  string `json:"username"`
	CSRFToken string `json:"csrfToken"`
}

type SessionResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username,omitempty"`
	CSRFToken     string `json:"csrfToken,omitempty"`
}

type FileObject struct {
	ContentType string
	Size        int64
}
