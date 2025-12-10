package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/SoulStalker/subscribes_api/internal/domain"
	"github.com/SoulStalker/subscribes_api/pkg/testutil"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestSubscriptionRepository_Create(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	sub := testutil.FixtureSubscription()

	rows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
		AddRow(sub.ID, sub.CreatedAt, sub.UpdatedAt)

	mock.ExpectQuery("INSERT INTO subscriptions").
		WithArgs(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		WillReturnRows(rows)
	err = repo.Create(ctx, sub)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Create_Error(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	sub := testutil.FixtureSubscription()

	mock.ExpectQuery("INSERT INTO subscriptions").
		WithArgs(sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate).
		WillReturnError(assert.AnError)

	err = repo.Create(ctx, sub)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_GetByID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	expectedSub := testutil.FixtureSubscription()

	rows := pgxmock.NewRows([]string{
		"id", "service_name", "price", "user_id",
		"start_date", "end_date", "created_at", "updated_at",
	}).AddRow(
		expectedSub.ID, expectedSub.ServiceName, expectedSub.Price, expectedSub.UserID,
		expectedSub.StartDate, expectedSub.EndDate, expectedSub.CreatedAt, expectedSub.UpdatedAt,
	)

	mock.ExpectQuery("SELECT (.+) FROM subscriptions WHERE id").
		WithArgs(expectedSub.ID).
		WillReturnRows(rows)

	sub, err := repo.GetByID(ctx, expectedSub.ID)

	require.NoError(t, err)
	assert.Equal(t, expectedSub.ID, sub.ID)
	assert.Equal(t, expectedSub.ServiceName, sub.ServiceName)
	assert.Equal(t, expectedSub.Price, sub.Price)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_List(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	userID := testutil.FixtureUserID()

	sub1 := testutil.FixtureSubscription(testutil.WithUserID(userID))
	sub2 := testutil.FixtureSubscription(testutil.WithUserID(userID))

	rows := pgxmock.NewRows([]string{
		"id", "service_name", "price", "user_id",
		"start_date", "end_date", "created_at", "updated_at",
	}).
		AddRow(sub1.ID, sub1.ServiceName, sub1.Price, sub1.UserID,
			sub1.StartDate, sub1.EndDate, sub1.CreatedAt, sub1.UpdatedAt).
		AddRow(sub2.ID, sub2.ServiceName, sub2.Price, sub2.UserID,
			sub2.StartDate, sub2.EndDate, sub2.CreatedAt, sub2.UpdatedAt)

	filter := domain.SubscriptionFilter{
		UserID: &userID,
	}

	mock.ExpectQuery("SELECT (.+) FROM subscriptions WHERE 1=1 AND user_id").
		WithArgs(userID).
		WillReturnRows(rows)

	subs, err := repo.List(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, subs, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Update(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	sub := testutil.FixtureSubscription()
	newUpdatedAt := time.Now()

	rows := pgxmock.NewRows([]string{"updated_at"}).AddRow(newUpdatedAt)

	mock.ExpectQuery("UPDATE subscriptions SET").
		WithArgs(sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.ID).
		WillReturnRows(rows)

	err = repo.Update(ctx, sub)

	require.NoError(t, err)
	assert.Equal(t, newUpdatedAt, sub.UpdatedAt)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_Delete(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	id := testutil.FixtureSubscriptionID()

	mock.ExpectExec("DELETE FROM subscriptions WHERE id").
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	err = repo.Delete(ctx, id)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSubscriptionRepository_CalculateTotalCost(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	logger := zaptest.NewLogger(t)
	repo := NewSubscriptionRepository(mock, logger)

	ctx := context.Background()
	startPeriod := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endPeriod := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	filter := domain.SubscriptionFilter{
		StartPeriod: &startPeriod,
		EndPeriod:   &endPeriod,
	}

	expectedTotal := 6000

	rows := pgxmock.NewRows([]string{"total"}).AddRow(expectedTotal)

	mock.ExpectQuery("SELECT COALESCE").
		WithArgs(startPeriod, endPeriod).
		WillReturnRows(rows)

	total, err := repo.TotalCost(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.NoError(t, mock.ExpectationsWereMet())
}
