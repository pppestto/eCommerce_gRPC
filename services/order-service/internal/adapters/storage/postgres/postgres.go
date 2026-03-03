package postgres

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/usecase"
)

const defaultDSN = "postgres://postgres:password@127.0.0.1:5433/ecommerce?sslmode=disable"

type PostgresRepository struct {
	db *pgxpool.Pool
}

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

func (r *PostgresRepository) SaveOrderWithOutbox(ctx context.Context, order *domain.Order, event usecase.OutboxEventSpec) error {
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

	outboxQuery := `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, outboxQuery,
		event.AggregateType(), event.AggregateID(), event.EventType(), event.Payload())
	if err != nil {
		return errors.Wrap(err, "failed to save outbox event")
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) UpdateStatusWithOutbox(ctx context.Context, id string, status domain.OrderStatus, event usecase.OutboxEventSpec) (*domain.Order, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin tx")
	}
	defer tx.Rollback(ctx)

	updateQuery := `
		UPDATE orders SET status = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	result, err := tx.Exec(ctx, updateQuery, int(status), id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update order status")
	}
	if result.RowsAffected() == 0 {
		return nil, commonerrors.ErrNotFound
	}

	outboxQuery := `
		INSERT INTO outbox (aggregate_type, aggregate_id, event_type, payload)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(ctx, outboxQuery,
		event.AggregateType(), event.AggregateID(), event.EventType(), event.Payload())
	if err != nil {
		return nil, errors.Wrap(err, "failed to save outbox event")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetByID(ctx, id)
}

func (r *PostgresRepository) GetUnpublishedOutbox(ctx context.Context, limit int) ([]OutboxRow, error) {
	query := `
		SELECT id::text, aggregate_type, aggregate_id, event_type, payload, created_at
		FROM outbox
		WHERE published_at IS NULL
		ORDER BY created_at
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query outbox")
	}
	defer rows.Close()

	var result []OutboxRow
	for rows.Next() {
		var row OutboxRow
		if err := rows.Scan(&row.ID, &row.AggregateType, &row.AggregateID, &row.EventType, &row.Payload, &row.CreatedAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan outbox row")
		}
		result = append(result, row)
	}
	return result, nil
}

type OutboxRow struct {
	ID            string
	AggregateType string
	AggregateID   string
	EventType     string
	Payload       json.RawMessage
	CreatedAt     time.Time
}

func (r *PostgresRepository) MarkOutboxPublished(ctx context.Context, id string) error {
	query := `UPDATE outbox SET published_at = now() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *PostgresRepository) IsProcessed(ctx context.Context, idempotencyKey string) (bool, error) {
	query := `SELECT 1 FROM processed_events WHERE idempotency_key = $1`
	var exists int
	err := r.db.QueryRow(ctx, query, idempotencyKey).Scan(&exists)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "failed to check processed_events")
	}
	return true, nil
}

func (r *PostgresRepository) MarkProcessed(ctx context.Context, idempotencyKey string) error {
	query := `INSERT INTO processed_events (idempotency_key) VALUES ($1) ON CONFLICT (idempotency_key) DO NOTHING`
	_, err := r.db.Exec(ctx, query, idempotencyKey)
	return err
}

func (r *PostgresRepository) LogOrderEvent(ctx context.Context, orderID, eventType string, payload json.RawMessage) error {
	query := `INSERT INTO order_events_log (order_id, event_type, payload) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, orderID, eventType, payload)
	return err
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
