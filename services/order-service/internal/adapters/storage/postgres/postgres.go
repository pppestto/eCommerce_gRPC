package postgres

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

const defaultDSN = "postgres://postgres:password@127.0.0.1:5433/ecommerce?sslmode=disable"

type PostgresRepository struct {
	db *pgxpool.Pool
}

// New создаёт новый postgres repository
func New(ctx context.Context) (*PostgresRepository, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDSN
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pool")
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	return &PostgresRepository{db: pool}, nil
}

func (r *PostgresRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin tx")
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders (id, user_id, total_amount, total_currency, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, orderQuery,
		order.ID, order.UserID, order.Total.Amount, order.Total.Currency, int(order.Status))
	if err != nil {
		return errors.Wrap(err, "failed to save order")
	}

	itemQuery := `
		INSERT INTO order_items (order_id, product_id, quantity, price_amount, price_currency)
		VALUES ($1, $2, $3, $4, $5)
	`
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, itemQuery,
			order.ID, item.ProductID, item.Quantity, item.Price.Amount, item.Price.Currency)
		if err != nil {
			return errors.Wrap(err, "failed to save order item")
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	orderQuery := `
		SELECT id, user_id, total_amount, total_currency, status
		FROM orders
		WHERE id = $1
	`
	var order domain.Order
	err := r.db.QueryRow(ctx, orderQuery, id).Scan(
		&order.ID, &order.UserID,
		&order.Total.Amount, &order.Total.Currency,
		&order.Status,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, commonerrors.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get order")
	}

	itemsQuery := `
		SELECT product_id, quantity, price_amount, price_currency
		FROM order_items
		WHERE order_id = $1
	`
	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order items")
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(&item.ProductID, &item.Quantity, &item.Price.Amount, &item.Price.Currency)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan order item")
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error) {
	query := `
		UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	result, err := r.db.Exec(ctx, query, int(status), id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update order status")
	}
	if result.RowsAffected() == 0 {
		return nil, commonerrors.ErrNotFound
	}

	return r.GetByID(ctx, id)
}
