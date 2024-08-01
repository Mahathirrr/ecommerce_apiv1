package storer

import "time"

type Product struct {
	ID           int        `db:"id"`
	Name         string     `db:"name"`
	Image        string     `db:"image"`
	Category     string     `db:"category"`
	Description  string     `db:"description"`
	Rating       int        `db:"rating"`
	NumReviews   int        `db:"num_reviews"`
	Price        float64    `db:"price"`
	CountInStock int        `db:"count_in_stock"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at"`
}

type Order struct {
	ID            int     `db:"id"`
	PaymentMethod string  `db:"payment_method"`
	TaxPrice      float64 `db:"tax_price"`
	ShippingPrice float64 `db:"shipping_price"`
	TotalPrice    float64 `db:"total_price"`
	UserId        int     `db:"user_id"`
	Items         []OrderItem
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
}

type OrderItem struct {
	ID        int     `db:"id"`
	Name      string  `db:"name"`
	Quantity  int     `db:"quantity"`
	Image     string  `db:"image"`
	Price     float64 `db:"price"`
	ProductID int     `db:"product_id"`
	OrderID   int     `db:"order_id"`
}

type User struct {
	ID        int        `db:"id"`
	Name      string     `db:"name"`
	Email     string     `db:"email"`
	Password  string     `db:"password"`
	IsAdmin   bool       `db:"is_admin"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type Session struct {
	ID           string    `db:"id"`
	UserEmail    string    `db:"user_email"`
	RefreshToken string    `db:"refresh_token"`
	IsRevoked    bool      `db:"is_revoked"`
	CreatedAt    time.Time `db:"created_at"`
	ExpiresAt    time.Time `db:"expires_at"`
}
