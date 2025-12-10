package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SoulStalker/subscribes_api/internal/domain"
	"github.com/SoulStalker/subscribes_api/pkg/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"
)

func TestSubscriptionService_CreateSubscription_Success(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	sub := testutil.FixtureSubscription()

	mockRepo.On("Create", ctx, sub).Return(nil)

	err := service.Create(ctx, sub)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_CreateSubscription_NegativePrice(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	sub := testutil.FixtureSubscription(testutil.WithPrice(-100))

	err := service.Create(ctx, sub)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "price cannot be negative")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestSubscriptionService_CreateSubscription_InvalidDateRange(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	startDate := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	sub := testutil.FixtureSubscription(testutil.WithDates(startDate, endDate))

	err := service.Create(ctx, sub)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end_date must be after start_date")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestSubscriptionService_GetSubscription_Success(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	expectedSub := testutil.FixtureSubscription()

	mockRepo.On("GetByID", ctx, expectedSub.ID).Return(expectedSub, nil)

	sub, err := service.GetByID(ctx, expectedSub.ID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSub, sub)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_GetSubscription_NotFound(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("GetByID", ctx, id).Return(nil, errors.New("not found"))

	sub, err := service.GetByID(ctx, id)

	assert.Error(t, err)
	assert.Nil(t, sub)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_ListSubscriptions(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	userID := testutil.FixtureUserID()
	filter := domain.SubscriptionFilter{UserID: &userID}

	expectedSubs := []domain.Subscription{
		*testutil.FixtureSubscription(testutil.WithUserID(userID)),
		*testutil.FixtureSubscription(testutil.WithUserID(userID)),
	}

	mockRepo.On("List", ctx, filter).Return(expectedSubs, nil)

	subs, err := service.List(ctx, filter)

	assert.NoError(t, err)
	assert.Len(t, subs, 2)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_UpdateSubscription_Success(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	existingSub := testutil.FixtureSubscription()
	updatedSub := testutil.FixtureSubscription(testutil.WithPrice(1000))
	updatedSub.ID = existingSub.ID

	mockRepo.On("GetByID", ctx, existingSub.ID).Return(existingSub, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Subscription")).Return(nil)

	err := service.Update(ctx, updatedSub)

	assert.NoError(t, err)
	assert.Equal(t, existingSub.CreatedAt, updatedSub.CreatedAt) // CreatedAt сохраняется
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_UpdateSubscription_NotFound(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	sub := testutil.FixtureSubscription()

	mockRepo.On("GetByID", ctx, sub.ID).Return(nil, errors.New("not found"))

	err := service.Update(ctx, sub)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subscription not found")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestSubscriptionService_DeleteSubscription(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	id := testutil.FixtureSubscriptionID()

	mockRepo.On("Delete", ctx, id).Return(nil)

	err := service.Delete(ctx, id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_TotalCost_Success(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	startPeriod := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endPeriod := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	filter := domain.SubscriptionFilter{
		StartPeriod: &startPeriod,
		EndPeriod:   &endPeriod,
	}

	expectedTotal := 5000

	mockRepo.On(
        "TotalCost",
        mock.Anything,
        mock.MatchedBy(func(f domain.SubscriptionFilter) bool {
            return f.StartPeriod != nil && f.EndPeriod != nil &&
                f.StartPeriod.Equal(startPeriod) && f.EndPeriod.Equal(endPeriod)
        }),
    ).Return(expectedTotal, nil)

	total, err := service.TotalCost(ctx, filter)

	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	mockRepo.AssertExpectations(t)
}

func TestSubscriptionService_TotalCost_MissingPeriods(t *testing.T) {
	mockRepo := new(testutil.MockSubscriptionRepository)
	logger := zaptest.NewLogger(t)
	service := NewSubscriptionService(mockRepo, logger)

	ctx := context.Background()
	filter := domain.SubscriptionFilter{} // Без дат

	total, err := service.TotalCost(ctx, filter)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start_period and end_period are required")
	assert.Zero(t, total)
	mockRepo.AssertNotCalled(t, "TotalCost")
}
