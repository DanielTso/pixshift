package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// User represents a row in the users table.
type User struct {
	ID                   string
	Email                string
	PasswordHash         sql.NullString
	Name                 string
	Provider             string
	Tier                 string
	StripeCustomerID     sql.NullString
	StripeSubscriptionID sql.NullString
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// CreateUser inserts a new user and returns the created record.
func (d *DB) CreateUser(ctx context.Context, email, passwordHash, name, provider string) (*User, error) {
	u := &User{}
	var pwHash sql.NullString
	if passwordHash != "" {
		pwHash = sql.NullString{String: passwordHash, Valid: true}
	}
	err := d.QueryRowContext(ctx,
		`INSERT INTO users (email, password_hash, name, provider)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, email, password_hash, name, provider, tier,
		           stripe_customer_id, stripe_subscription_id, created_at, updated_at`,
		email, pwHash, name, provider,
	).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Provider, &u.Tier,
		&u.StripeCustomerID, &u.StripeSubscriptionID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetUserByEmail returns the user with the given email.
func (d *DB) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := &User{}
	err := d.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, provider, tier,
		        stripe_customer_id, stripe_subscription_id, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Provider, &u.Tier,
		&u.StripeCustomerID, &u.StripeSubscriptionID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// GetUserByID returns the user with the given ID.
func (d *DB) GetUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := d.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, provider, tier,
		        stripe_customer_id, stripe_subscription_id, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Provider, &u.Tier,
		&u.StripeCustomerID, &u.StripeSubscriptionID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// GetUserByStripeCustomer returns the user with the given Stripe customer ID.
func (d *DB) GetUserByStripeCustomer(ctx context.Context, customerID string) (*User, error) {
	u := &User{}
	err := d.QueryRowContext(ctx,
		`SELECT id, email, password_hash, name, provider, tier,
		        stripe_customer_id, stripe_subscription_id, created_at, updated_at
		 FROM users WHERE stripe_customer_id = $1`, customerID,
	).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Provider, &u.Tier,
		&u.StripeCustomerID, &u.StripeSubscriptionID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by stripe customer: %w", err)
	}
	return u, nil
}

// UpdateUserTier sets the tier for a user.
func (d *DB) UpdateUserTier(ctx context.Context, userID, tier string) error {
	_, err := d.ExecContext(ctx,
		`UPDATE users SET tier = $1, updated_at = now() WHERE id = $2`,
		tier, userID,
	)
	if err != nil {
		return fmt.Errorf("update user tier: %w", err)
	}
	return nil
}

// UpdateStripeCustomer sets the stripe_customer_id for a user.
func (d *DB) UpdateStripeCustomer(ctx context.Context, userID, customerID string) error {
	_, err := d.ExecContext(ctx,
		`UPDATE users SET stripe_customer_id = $1, updated_at = now() WHERE id = $2`,
		customerID, userID,
	)
	if err != nil {
		return fmt.Errorf("update stripe customer: %w", err)
	}
	return nil
}

// UpdateStripeSubscription sets the stripe_subscription_id for a user.
func (d *DB) UpdateStripeSubscription(ctx context.Context, userID, subscriptionID string) error {
	_, err := d.ExecContext(ctx,
		`UPDATE users SET stripe_subscription_id = $1, updated_at = now() WHERE id = $2`,
		subscriptionID, userID,
	)
	if err != nil {
		return fmt.Errorf("update stripe subscription: %w", err)
	}
	return nil
}
