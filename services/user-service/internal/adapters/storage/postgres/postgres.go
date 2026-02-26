package postgres

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
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
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresRepository{db: pool}, nil
}

// Save сохраняет пользователя в БД
func (r *PostgresRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email) 
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET email = $2
	`

	_, err := r.db.Exec(ctx, query, user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

// GetByID получает пользователя по ID
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := "SELECT id, email FROM users WHERE id = $1"

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

// Delete удаляет пользователя по ID
func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Close закрывает соединение с БД
func (r *PostgresRepository) Close() {
	r.db.Close()
}
