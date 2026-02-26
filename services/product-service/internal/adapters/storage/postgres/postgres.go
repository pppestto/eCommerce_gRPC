package postgres

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
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

func (r *PostgresRepository) Save(ctx context.Context, product *domain.Product) error {
	query := `INSERT INTO products (id, name, description, price_amount, price_currency, stock)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		product.ID, product.Name, product.Description,
		product.Price.Amount, product.Price.Currency, product.Stock)
	if err != nil {
		return errors.Wrap(err, "failed to save product")
	}

	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `SELECT id, name, description, price_amount, price_currency, stock
		FROM products WHERE id = $1`

	var product domain.Product
	err := r.db.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Description,
		&product.Price.Amount, &product.Price.Currency, &product.Stock,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, commonerrors.ErrNotFound
		}
		return nil, errors.Wrap(err, "failed to get product")
	}

	return &product, nil
}

// List возвращает продукты с пагинацией и опциональной фильтрацией по категории.
// page, size — пагинация (page начиная с 1); category — пустая строка = без фильтра.
func (r *PostgresRepository) List(ctx context.Context, page, size int, category string) ([]*domain.Product, int, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	// Считаем total
	countQuery := `SELECT COUNT(*) FROM products`
	countArgs := []interface{}{}
	if category != "" {
		countQuery += ` WHERE category = $1`
		countArgs = append(countArgs, category)
	}

	var total int
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to count products")
	}

	// Выбираем продукты
	var listQuery string
	var args []interface{}
	if category != "" {
		listQuery = `SELECT id, name, description, price_amount, price_currency, stock
			FROM products WHERE category = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
		args = []interface{}{category, size, offset}
	} else {
		listQuery = `SELECT id, name, description, price_amount, price_currency, stock
			FROM products ORDER BY created_at DESC LIMIT $1 OFFSET $2`
		args = []interface{}{size, offset}
	}

	rows, err := r.db.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list products")
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		var p domain.Product
		err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price.Amount, &p.Price.Currency, &p.Stock)
		if err != nil {
			return nil, 0, errors.Wrap(err, "failed to scan product")
		}
		products = append(products, &p)
	}

	return products, total, nil
}
