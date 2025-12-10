package testutil

import (
	"time"

	"github.com/SoulStalker/subscribes_api/internal/domain"
	"github.com/google/uuid"
)

// FixtureSubscription создает тестовую подписку
func FixtureSubscription(opts ...func(*domain.Subscription)) *domain.Subscription {
	now := time.Now()
	startDate := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)

	sub := &domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Test Service",
		Price:       500,
		UserID:      uuid.New(),
		StartDate:   startDate,
		EndDate:     nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for _, opt := range opts {
		opt(sub)
	}

	return sub
}

// WithServiceName устанавливает название сервиса
func WithServiceName(name string) func(*domain.Subscription) {
	return func(s *domain.Subscription) {
		s.ServiceName = name
	}
}

// WithPrice устанавливает цену
func WithPrice(price int) func(*domain.Subscription) {
	return func(s *domain.Subscription) {
		s.Price = price
	}
}

// WithUserID устанавливает user_id
func WithUserID(userID uuid.UUID) func(*domain.Subscription) {
	return func(s *domain.Subscription) {
		s.UserID = userID
	}
}

// WithDates устанавливает даты
func WithDates(start, end time.Time) func(*domain.Subscription) {
	return func(s *domain.Subscription) {
		s.StartDate = start
		if !end.IsZero() {
			s.EndDate = &end
		}
	}
}

// WithEndDate устанавливает дату окончания
func WithEndDate(end *time.Time) func(*domain.Subscription) {
	return func(s *domain.Subscription) {
		s.EndDate = end
	}
}

// FixtureUserID возвращает фиксированный UUID для тестов
func FixtureUserID() uuid.UUID {
	return uuid.MustParse("60601fee-2bf1-4721-ae6f-7636e79a0cba")
}

// FixtureSubscriptionID возвращает фиксированный ID подписки
func FixtureSubscriptionID() uuid.UUID {
	return uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
}
