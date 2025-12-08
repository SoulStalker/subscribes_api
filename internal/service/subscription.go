package service

import (
	"context"
	"fmt"

	"github.com/SoulStalker/subscribes_api/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	TotalCost(ctx context.Context, filter domain.SubscriptionFilter) (int, error)
}

type SubscriptionService struct {
	repo   SubscriptionRepository
	logger *zap.Logger
}

func NewSubscriptionService(repo SubscriptionRepository, logger *zap.Logger) *SubscriptionService {
	return &SubscriptionService{
		repo:   repo,
		logger: logger,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *domain.Subscription) error {
	if sub.Price < 0 {
		return fmt.Errorf("price cannot be negative")
	}

	if sub.EndDate != nil && sub.EndDate.Before(sub.StartDate) {
		return fmt.Errorf("end_date must be after start_date")
	}

	return s.repo.Create(ctx, sub)
}

func (s *SubscriptionService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *SubscriptionService) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	return s.repo.List(ctx, filter)
}

func (s *SubscriptionService) Update(ctx context.Context, sub *domain.Subscription) error {
	existing, err := s.repo.GetByID(ctx, sub.ID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	sub.CreatedAt = existing.CreatedAt
	return s.repo.Update(ctx, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *SubscriptionService) TotalCost(ctx context.Context, filter domain.SubscriptionFilter) (int, error) {
	if filter.StartPeriod == nil || filter.EndPeriod == nil {
		return 0, fmt.Errorf("start_period and end_period are required")
	}

	return s.repo.TotalCost(ctx, filter)
}
