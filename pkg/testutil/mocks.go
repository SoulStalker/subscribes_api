package testutil

import (
	"context"

	"github.com/SoulStalker/subscribes_api/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockSubscriptionRepository мок репозитория
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) List(ctx context.Context, filter domain.SubscriptionFilter) ([]domain.Subscription, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]domain.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) TotalCost(ctx context.Context, filter domain.SubscriptionFilter) (int, error) {
	args := m.Called(ctx, filter)
	return args.Int(0), args.Error(1)
}
