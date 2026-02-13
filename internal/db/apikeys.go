package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// APIKey represents a row in the api_keys table.
type APIKey struct {
	ID        string
	UserID    string
	KeyHash   string
	Prefix    string
	Name      string
	RevokedAt sql.NullTime
	CreatedAt time.Time
}

// CreateAPIKey inserts a new API key record.
func (d *DB) CreateAPIKey(ctx context.Context, userID, keyHash, prefix, name string) (*APIKey, error) {
	k := &APIKey{}
	err := d.QueryRowContext(ctx,
		`INSERT INTO api_keys (user_id, key_hash, prefix, name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, key_hash, prefix, name, revoked_at, created_at`,
		userID, keyHash, prefix, name,
	).Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Prefix, &k.Name, &k.RevokedAt, &k.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create api key: %w", err)
	}
	return k, nil
}

// GetAPIKeyByHash looks up an API key by its SHA-256 hash and returns the key
// along with the owning user. Only non-revoked keys are returned.
func (d *DB) GetAPIKeyByHash(ctx context.Context, hash string) (*APIKey, *User, error) {
	k := &APIKey{}
	u := &User{}
	err := d.QueryRowContext(ctx,
		`SELECT k.id, k.user_id, k.key_hash, k.prefix, k.name, k.revoked_at, k.created_at,
		        u.id, u.email, u.password_hash, u.name, u.provider, u.tier,
		        u.stripe_customer_id, u.stripe_subscription_id, u.created_at, u.updated_at
		 FROM api_keys k
		 JOIN users u ON u.id = k.user_id
		 WHERE k.key_hash = $1 AND k.revoked_at IS NULL`, hash,
	).Scan(
		&k.ID, &k.UserID, &k.KeyHash, &k.Prefix, &k.Name, &k.RevokedAt, &k.CreatedAt,
		&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Provider, &u.Tier,
		&u.StripeCustomerID, &u.StripeSubscriptionID, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("get api key by hash: %w", err)
	}
	return k, u, nil
}

// ListAPIKeys returns all non-revoked API keys for a user.
func (d *DB) ListAPIKeys(ctx context.Context, userID string) ([]APIKey, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT id, user_id, key_hash, prefix, name, revoked_at, created_at
		 FROM api_keys
		 WHERE user_id = $1 AND revoked_at IS NULL
		 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var keys []APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.KeyHash, &k.Prefix, &k.Name, &k.RevokedAt, &k.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

// RevokeAPIKey marks an API key as revoked. The key must belong to the given user.
func (d *DB) RevokeAPIKey(ctx context.Context, keyID, userID string) error {
	_, err := d.ExecContext(ctx,
		`UPDATE api_keys SET revoked_at = now() WHERE id = $1 AND user_id = $2`,
		keyID, userID,
	)
	if err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}

// CountActiveKeys returns the number of non-revoked API keys for a user.
func (d *DB) CountActiveKeys(ctx context.Context, userID string) (int, error) {
	var count int
	err := d.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM api_keys WHERE user_id = $1 AND revoked_at IS NULL`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active keys: %w", err)
	}
	return count, nil
}
