package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Conversion represents a row in the conversions table.
type Conversion struct {
	ID           string
	UserID       sql.NullString
	APIKeyID     sql.NullString
	InputFormat  string
	OutputFormat string
	InputSize    int64
	OutputSize   int64
	DurationMS   int
	Params       json.RawMessage
	Source       string
	CreatedAt    time.Time
}

// RecordConversion inserts a conversion record.
func (d *DB) RecordConversion(ctx context.Context, c *Conversion) error {
	_, err := d.ExecContext(ctx,
		`INSERT INTO conversions (user_id, api_key_id, input_format, output_format,
		                          input_size, output_size, duration_ms, params, source)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		c.UserID, c.APIKeyID, c.InputFormat, c.OutputFormat,
		c.InputSize, c.OutputSize, c.DurationMS, c.Params, c.Source,
	)
	if err != nil {
		return fmt.Errorf("record conversion: %w", err)
	}
	return nil
}

// ListConversions returns conversions for a user, ordered by most recent first.
func (d *DB) ListConversions(ctx context.Context, userID string, limit, offset int) ([]Conversion, error) {
	rows, err := d.QueryContext(ctx,
		`SELECT id, user_id, api_key_id, input_format, output_format,
		        input_size, output_size, duration_ms, params, source, created_at
		 FROM conversions
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`, userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list conversions: %w", err)
	}
	defer rows.Close()

	var convs []Conversion
	for rows.Next() {
		var c Conversion
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.APIKeyID, &c.InputFormat, &c.OutputFormat,
			&c.InputSize, &c.OutputSize, &c.DurationMS, &c.Params, &c.Source, &c.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan conversion: %w", err)
		}
		convs = append(convs, c)
	}
	return convs, rows.Err()
}
