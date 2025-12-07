package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/SoulStalker/subscribes_api/internal/domain"
)

type SubscriptionRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewSubscriptionRepository(db *pgxpool.Pool, logger *zap.Logger) *SubscriptionRepository {
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
		return fmt.Errorf("create subscrioption: %w", &err)
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
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("subsription not found", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	return &sub, nil
}

func (r *SubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	query := `SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at FROM subscriptions WHERE 1=1`
	args := []any{}
	argID := 1

	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = &id", argID)
		args = append(args, &filter.UserID)
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
