package db

import (
	"context"
	"database/sql"
	"fmt"
)

// IncrementUsage atomically increments the daily conversion count for a user
// and returns the new count.
func (d *DB) IncrementUsage(ctx context.Context, userID string) (int, error) {
	var count int
	err := d.QueryRowContext(ctx,
		`INSERT INTO daily_usage (user_id, date, count)
		 VALUES ($1, CURRENT_DATE, 1)
		 ON CONFLICT (user_id, date) DO UPDATE SET count = daily_usage.count + 1
		 RETURNING count`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("increment usage: %w", err)
	}
	return count, nil
}

// GetDailyUsage returns the current day's conversion count for a user.
// Returns 0 if no record exists for today.
func (d *DB) GetDailyUsage(ctx context.Context, userID string) (int, error) {
	var count int
	err := d.QueryRowContext(ctx,
		`SELECT count FROM daily_usage WHERE user_id = $1 AND date = CURRENT_DATE`,
		userID,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get daily usage: %w", err)
	}
	return count, nil
}

// GetMonthlyAPIUsage returns the current month's API request count for a user.
// Returns 0 if no record exists for this month.
func (d *DB) GetMonthlyAPIUsage(ctx context.Context, userID string) (int, error) {
	var count int
	err := d.QueryRowContext(ctx,
		`SELECT count FROM monthly_api_usage WHERE user_id = $1 AND month = date_trunc('month', CURRENT_DATE)::date`,
		userID,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get monthly api usage: %w", err)
	}
	return count, nil
}

// IncrementMonthlyAPIUsage atomically increments the monthly API request count
// for a user and returns the new count.
func (d *DB) IncrementMonthlyAPIUsage(ctx context.Context, userID string) (int, error) {
	var count int
	err := d.QueryRowContext(ctx,
		`INSERT INTO monthly_api_usage (user_id, month, count)
		 VALUES ($1, date_trunc('month', CURRENT_DATE)::date, 1)
		 ON CONFLICT (user_id, month) DO UPDATE SET count = monthly_api_usage.count + 1
		 RETURNING count`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("increment monthly api usage: %w", err)
	}
	return count, nil
}
