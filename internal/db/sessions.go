package db

import (
	"context"
	"fmt"
	"time"
)

// Session represents a row in the sessions table.
type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// CreateSession inserts a new session.
func (d *DB) CreateSession(ctx context.Context, userID, token string, expiresAt time.Time) (*Session, error) {
	s := &Session{}
	err := d.QueryRowContext(ctx,
		`INSERT INTO sessions (id, user_id, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, expires_at, created_at`,
		token, userID, expiresAt,
	).Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return s, nil
}

// GetSession returns the session and its associated user. Returns an error if
// the session does not exist or has expired.
func (d *DB) GetSession(ctx context.Context, token string) (*Session, *User, error) {
	s := &Session{}
	u := &User{}
	err := d.QueryRowContext(ctx,
		`SELECT s.id, s.user_id, s.expires_at, s.created_at,
		        u.id, u.email, u.password_hash, u.name, u.provider, u.tier,
		        u.stripe_customer_id, u.stripe_subscription_id, u.created_at, u.updated_at
		 FROM sessions s
		 JOIN users u ON u.id = s.user_id
		 WHERE s.id = $1 AND s.expires_at > now()`, token,
	).Scan(
		&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt,
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Provider, &u.Tier,
		&u.StripeCustomerID, &u.StripeSubscriptionID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("get session: %w", err)
	}
	return s, u, nil
}

// DeleteSession removes a session by token.
func (d *DB) DeleteSession(ctx context.Context, token string) error {
	_, err := d.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// DeleteExpiredSessions removes all expired sessions and returns the count deleted.
func (d *DB) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	res, err := d.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= now()`)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}
	return res.RowsAffected()
}
