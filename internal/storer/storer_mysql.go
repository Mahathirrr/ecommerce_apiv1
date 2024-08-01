package storer

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type MySQLStorage struct {
	*sqlx.DB
}

func NewMySQLStorage(db *sqlx.DB) *MySQLStorage {
	return &MySQLStorage{
		DB: db,
	}
}

func (ms *MySQLStorage) CreateProduct(ctx context.Context, p *Product) (*Product, error) {
	res, err := ms.DB.NamedExecContext(ctx, "INSERT INTO products (name, image, category, description, rating, num_reviews, price, count_in_stock) VALUES (:name, :image, :category, :description, :rating, :num_reviews, :price, :count_in_stock)", p)
	if err != nil {
		return nil, fmt.Errorf("error inserting product: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting last insert ID: %w", err)
	}
	p.ID = int(id)
	return p, nil
}

func (ms *MySQLStorage) GetProduct(ctx context.Context, id int) (*Product, error) {
	var p Product
	err := ms.GetContext(ctx, &p, "SELECT * FROM products WHERE ID = ?", id)
	if err != nil {
		return nil, fmt.Errorf("error getting product: %w", err)
	}
	return &p, nil
}

func (ms *MySQLStorage) ListProducts(ctx context.Context) ([]Product, error) {
	var products []Product
	err := ms.DB.SelectContext(ctx, &products, "SELECT * FROM products")
	if err != nil {
		return nil, fmt.Errorf("error listing products: %w", err)
	}
	return products, nil
}

func (ms *MySQLStorage) UpdateProduct(ctx context.Context, p *Product) (*Product, error) {
	_, err := ms.DB.NamedExecContext(ctx, "UPDATE products SET name=:name, image=:image, category=:category, description=:description, rating=:rating, num_reviews=:num_reviews, price=:price, count_in_stock=:count_in_stock, updated_at=:updated_at WHERE id=:id", p)
	if err != nil {
		return nil, fmt.Errorf("error updating product: %w", err)
	}
	return p, nil
}

func (ms *MySQLStorage) DeleteProduct(ctx context.Context, id int) error {
	_, err := ms.DB.ExecContext(ctx, "DELETE FROM products where id=?", id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}
	return nil
}

func (ms *MySQLStorage) execTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := ms.DB.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %w", err)
		}
		return fmt.Errorf("error in transaction: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func createOrder(ctx context.Context, tx *sqlx.Tx, o *Order) (*Order, error) {
	res, err := tx.NamedExecContext(ctx, "INSERT INTO orders (payment_method, tax_price, shipping_price, total_price, user_id) VALUES (:payment_method, :tax_price, :shipping_price, :total_price, :user_id)", o)
	if err != nil {
		return nil, fmt.Errorf("error inserting order: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting last insert id order: %w", err)
	}
	o.ID = int(id)
	return o, nil
}

func createOrderItem(ctx context.Context, tx *sqlx.Tx, oi *OrderItem) error {
	res, err := tx.NamedExecContext(ctx, "INSERT INTO order_items (name, quantity, image, price, product_id, order_id) VALUES (:name, :quantity, :image, :price, :product_id, :order_id)", oi)
	if err != nil {
		return fmt.Errorf("error inserting order item: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert id order item: %w", err)
	}
	oi.ID = int(id)
	return nil
}

func (ms *MySQLStorage) CreateOrder(ctx context.Context, o *Order) (*Order, error) {
	err := ms.execTx(ctx, func(tx *sqlx.Tx) error {
		/*
		   result, err := tx.NamedExecContext(ctx, `
		       INSERT INTO orders (payment_method, tax_price, shipping_price, total_price, user_id, created_at, updated_at)
		       VALUES (:payment_method, :tax_price, :shipping_price, :total_price, :user_id, :created_at, :updated_at)
		   `, o)
		   if err != nil {
		       return fmt.Errorf("error inserting order: %w", err)
		   }

		   id, err := result.LastInsertId()
		   if err != nil {
		       return fmt.Errorf("error getting last insert ID for order: %w", err)
		   }
		   o.ID = int(id)

		   // Insert order items
		   for i := range o.Items {
		       o.Items[i].OrderID = o.ID
		       _, err = tx.NamedExecContext(ctx, `
		           INSERT INTO order_items (name, quantity, image, price, product_id, order_id)
		           VALUES (:name, :quantity, :image, :price, :product_id, :order_id)
		       `, o.Items[i])
		       if err != nil {
		           return fmt.Errorf("error inserting order item: %w", err)
		       }
		   }
		*/
		order, err := createOrder(ctx, tx, o)
		if err != nil {
			return fmt.Errorf("error creating order: %w", err)
		}

		for _, oi := range o.Items {
			oi.OrderID = order.ID
			// insert into order_items
			// pembeda dgn ecomm, &oi
			err = createOrderItem(ctx, tx, &oi)
			if err != nil {
				return fmt.Errorf("error creating order item: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error creating order: %w", err)
	}
	return o, nil
}

func (ms *MySQLStorage) GetOrder(ctx context.Context, userId int) (*Order, error) {
	var o Order
	err := ms.DB.GetContext(ctx, &o, "SELECT * FROM orders WHERE user_id=?", userId)
	if err != nil {
		return nil, fmt.Errorf("error getting order: %w", err)
	}
	var oi []OrderItem
	err = ms.DB.SelectContext(ctx, &oi, "SELECT * FROM order_items WHERE order_id=?", o.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting order item: %w", err)
	}
	o.Items = oi
	return &o, nil
}

func (ms *MySQLStorage) ListOrders(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := ms.DB.SelectContext(ctx, &orders, "SELECT * FROM orders")
	if err != nil {
		return nil, fmt.Errorf("error listing order: %w", err)
	}
	for i := range orders {
		var items []OrderItem
		err := ms.DB.SelectContext(ctx, &items, "SELECT * FROM order_items WHERE order_id=?", orders[i].ID)
		if err != nil {
			return nil, fmt.Errorf("error listing order items: %w", err)
		}
	}
	return orders, nil
}

func (ms *MySQLStorage) DeleteOrders(ctx context.Context, id int) error {
	err := ms.execTx(ctx, func(tx *sqlx.Tx) error {
		_, err := tx.ExecContext(ctx, "DELETE FROM order_items where order_id=?", id)
		if err != nil {
			return fmt.Errorf("error deleting orders: %w", err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM orders where id=?", id)
		if err != nil {
			return fmt.Errorf("error deleting order: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error deleting order: %w", err)
	}
	return nil
}

func (ms *MySQLStorage) CreateUser(ctx context.Context, u *User) (*User, error) {
	res, err := ms.DB.NamedExecContext(ctx, "INSERT INTO users (name, email, password, is_admin) VALUES (:name, :email, :password, :is_admin)", u)
	if err != nil {
		return nil, fmt.Errorf("error inserting user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error getting last insert ID: %w", err)
	}
	u.ID = int(id)

	return u, nil
}

func (ms *MySQLStorage) GetUser(ctx context.Context, email string) (*User, error) {
	var u User
	err := ms.DB.GetContext(ctx, &u, "SELECT * FROM users WHERE email=?", email)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %w", err)
	}

	return &u, nil
}

func (ms *MySQLStorage) ListUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := ms.DB.SelectContext(ctx, &users, "SELECT * FROM users")
	if err != nil {
		return nil, fmt.Errorf("error listing users: %w", err)
	}

	return users, nil
}

func (ms *MySQLStorage) UpdateUser(ctx context.Context, u *User) (*User, error) {
	_, err := ms.DB.NamedExecContext(ctx, "UPDATE users SET name=:name, email=:email, password=:password, is_admin=:is_admin, updated_at=:updated_at WHERE id=:id", u)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}

	return u, nil
}

func (ms *MySQLStorage) DeleteUser(ctx context.Context, id int) error {
	_, err := ms.DB.ExecContext(ctx, "DELETE FROM users WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

func (ms *MySQLStorage) CreateSession(ctx context.Context, s *Session) (*Session, error) {
	_, err := ms.DB.NamedExecContext(ctx, "INSERT INTO sessions (id, user_email, refresh_token, is_revoked, expires_at) VALUES (:id, :user_email, :refresh_token, :is_revoked, :expires_at)", s)
	if err != nil {
		return nil, fmt.Errorf("error inserting session: %w", err)
	}

	return s, nil
}

func (ms *MySQLStorage) GetSession(ctx context.Context, id string) (*Session, error) {
	var s Session
	err := ms.DB.GetContext(ctx, &s, "SELECT * FROM sessions WHERE id=?", id)
	if err != nil {
		return nil, fmt.Errorf("error getting session: %w", err)
	}

	return &s, nil
}

func (ms *MySQLStorage) RevokeSession(ctx context.Context, id string) error {
	// menggunakan map sebagai inline parameternya, karena namedExecContext
	// mengharapkan map/struct sebagai parameternya. jika ingin langsung menggunakan
	// id sebagai inline condition, dapat menggunakan execContext
	_, err := ms.DB.NamedExecContext(ctx, "UPDATE sessions SET is_revoked=1 WHERE id=:id", map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("error revoking session: %w", err)
	}

	return nil
}

func (ms *MySQLStorage) DeleteSession(ctx context.Context, id string) error {
	_, err := ms.DB.ExecContext(ctx, "DELETE FROM sessions WHERE id=?", id)
	if err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}
