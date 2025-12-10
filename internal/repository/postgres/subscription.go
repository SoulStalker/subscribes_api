package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/SoulStalker/subscribes_api/internal/domain"
)

// PgxPool интерфейс для возможности тестов через pgxmock
type PgxPool interface {
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

type SubscriptionRepository struct {
	db     PgxPool
	logger *zap.Logger
}

func NewSubscriptionRepository(db PgxPool, logger *zap.Logger) *SubscriptionRepository {
	return &SubscriptionRepository{db: db, logger: logger}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	query := `
		 INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at 
	`

	r.logger.Debug("creating subscription", zap.String("service", sub.ServiceName))

	err := r.db.QueryRow(ctx, query, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to create subscription", zap.Error(err))
		return fmt.Errorf("create subscription: %w", err)
	}

	r.logger.Info("subscription created", zap.String("id", sub.ID.String()))
	return nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	query := `
        SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
        FROM subscriptions
        WHERE id = $1
    `
	var sub domain.Subscription
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID,
		&sub.ServiceName,
		&sub.Price,
		&sub.UserID,
		&sub.StartDate,
		&sub.EndDate,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("subscription not found", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	return &sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE 1=1`
	args := []any{}
	argID := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argID)
		args = append(args, *filter.UserID)
		argID++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argID)
		args = append(args, "%"+*filter.ServiceName+"%")
		argID++
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("List subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	r.logger.Debug("subscriptions list", zap.Int("count", len(subs)))
	return subs, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	query := `
        UPDATE subscriptions
        SET service_name = $1, price = $2, start_date = $3, end_date = $4
        WHERE id = $5
        RETURNING updated_at
    `

	err := r.db.QueryRow(ctx, query, sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.ID).Scan(&sub.UpdatedAt)

	if err != nil {
		r.logger.Error("failed to update subscription", zap.String("id", sub.ID.String()), zap.Error(err))
		return fmt.Errorf("update subscription: %w", err)
	}

	r.logger.Info("subscription updated", zap.String("id", sub.ID.String()))
	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		r.logger.Error("failed to delete subscription", zap.String("id", id.String()), zap.Error(err))
		return fmt.Errorf("delete subscription")
	}

	r.logger.Info("subscription deleted", zap.String("id", id.String()))
	return nil
}

func (r *SubscriptionRepository) TotalCost(ctx context.Context, filter domain.SubscriptionFilter) (int, error) {
	query := `
	        SELECT COALESCE(SUM(
            price * (
                EXTRACT(YEAR FROM AGE(
                    LEAST(COALESCE(end_date, $2), $2),
                    GREATEST(start_date, $1)
                )) * 12 +
                EXTRACT(MONTH FROM AGE(
                    LEAST(COALESCE(end_date, $2), $2),
                    GREATEST(start_date, $1)
                )) + 1
            )
        ), 0)::INTEGER
        FROM subscriptions
        WHERE start_date <= $2
          AND (end_date IS NULL OR end_date >= $1)
	`

	args := []any{filter.StartPeriod, filter.EndPeriod}
	argID := 3

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argID)
		args = append(args, *filter.UserID)
		argID++
	}

	if filter.ServiceName != nil {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argID)
		args = append(args, "%"+*filter.ServiceName+"%")
	}

	var total int
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		r.logger.Error("failed to calculate total", zap.Error(err))
		return 0, fmt.Errorf("calculate total: %w", err)
	}

	r.logger.Info("total cost calculated", zap.Int("total", total))
	return total, nil
}
