package models

type Category struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type Product struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Price  int    `json:"price"`
	Cat    string `json:"cat"`
	Sub    string `json:"sub"`
	Img    string `json:"img,omitempty"`
	IsNew  bool   `json:"isNew,omitempty"`
	Size   string `json:"size"`
	Colors string `json:"colors"`
	Canvas string `json:"canvas"`
}

type GalleryItem struct {
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
	Name string `json:"name"`
	Role string `json:"role"`
	Img  string `json:"img"`
	Text string `json:"text"`
}

type SiteContent struct {
	Author       Author        `json:"author"`
	HowToBuy     []HowToStep   `json:"howToBuy"`
	Testimonials []Testimonial `json:"testimonials"`
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
