package usecase

import (
	"context"
	"time"

	"github.com/herpiko/blankon-telemetry-backend/internal/repo"
)

type AnalyticsUsecase interface {
	GetHourlyStats(ctx context.Context, eventName string, from, to time.Time) ([]repo.EventStats, error)
	GetDailyStats(ctx context.Context, eventName string, from, to time.Time) ([]repo.EventStats, error)
}

type analyticsUsecase struct {
	repo repo.AnalyticsRepository
}

func NewAnalyticsUsecase(repo repo.AnalyticsRepository) AnalyticsUsecase {
	return &analyticsUsecase{repo: repo}
}

func (u *analyticsUsecase) GetHourlyStats(ctx context.Context, eventName string, from, to time.Time) ([]repo.EventStats, error) {
	// Default to last 24 hours if not specified
	if from.IsZero() {
		from = time.Now().UTC().Add(-24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}

	return u.repo.GetHourlyStats(ctx, eventName, from, to)
}

func (u *analyticsUsecase) GetDailyStats(ctx context.Context, eventName string, from, to time.Time) ([]repo.EventStats, error) {
	// Default to last 30 days if not specified
	if from.IsZero() {
		from = time.Now().UTC().Add(-30 * 24 * time.Hour)
	}
	if to.IsZero() {
		to = time.Now().UTC()
	}

	return u.repo.GetDailyStats(ctx, eventName, from, to)
}
