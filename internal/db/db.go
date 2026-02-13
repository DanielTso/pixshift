package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps a *sql.DB with application-level helpers.
type DB struct {
	*sql.DB
}

// Open connects to Postgres using DATABASE_URL from the environment.
func Open() (*DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL not set")
	}
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &DB{conn}, nil
}

// OpenWithDSN connects to Postgres with an explicit DSN.
func OpenWithDSN(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return &DB{conn}, nil
}

// Migrate runs all SQL migration files in order.
func (d *DB) Migrate(ctx context.Context, migrationsSQL string) error {
	statements := splitStatements(migrationsSQL)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := d.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w\nstatement: %s", err, stmt)
		}
	}
	return nil
}

func splitStatements(sql string) []string {
	// Split on semicolons but respect dollar-quoted strings
	var stmts []string
	current := strings.Builder{}
	inDollarQuote := false
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		if strings.Contains(line, "$$") {
			inDollarQuote = !inDollarQuote
		}
		current.WriteString(line)
		current.WriteString("\n")
		if !inDollarQuote && strings.HasSuffix(trimmed, ";") {
			stmts = append(stmts, current.String())
			current.Reset()
		}
	}
	if s := current.String(); strings.TrimSpace(s) != "" {
		stmts = append(stmts, s)
	}
	return stmts
}
