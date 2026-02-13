package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/billing"
	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/db"
	"github.com/DanielTso/pixshift/internal/server"
	"github.com/DanielTso/pixshift/web"
)

// migrationSQL is embedded from the migrations directory.
const migrationSQL = "" // migrations are run via db.Migrate with the SQL file content

func runServeMode(ctx context.Context, registry *codec.Registry, opts *options) {
	fmt.Printf("Starting Pixshift HTTP server on %s...\n", opts.serveAddr)
	srv := server.New(registry, opts.serveAddr)

	// Simple mode settings (always applied)
	if opts.apiKey != "" {
		srv.APIKey = opts.apiKey
	}
	if opts.rateLimit > 0 {
		srv.RateLimit = opts.rateLimit
	}
	if opts.corsOrigins != "" {
		srv.AllowOrigins = opts.corsOrigins
	}
	if opts.requestTimeout > 0 {
		srv.Timeout = time.Duration(opts.requestTimeout) * time.Second
	}
	if opts.maxUpload > 0 {
		srv.MaxFileSize = opts.maxUpload
	}

	// Full mode: activate when DATABASE_URL is set
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		database, err := db.OpenWithDSN(dsn)
		if err != nil {
			fatal("database: %v", err)
		}
		defer database.Close()
		fmt.Fprintf(os.Stderr, "Connected to database\n")

		srv.DB = database

		// Session secret
		srv.SessionSecret = os.Getenv("SESSION_SECRET")
		if srv.SessionSecret == "" {
			fatal("SESSION_SECRET is required when DATABASE_URL is set")
		}

		// Base URL
		srv.BaseURL = os.Getenv("BASE_URL")
		if srv.BaseURL == "" {
			srv.BaseURL = "http://localhost" + opts.serveAddr
		}

		// Stripe billing
		if stripeKey := os.Getenv("STRIPE_SECRET_KEY"); stripeKey != "" {
			billing.Init(stripeKey)
			billing.ProMonthlyPriceID = os.Getenv("STRIPE_PRO_MONTHLY_PRICE_ID")
			billing.ProAnnualPriceID = os.Getenv("STRIPE_PRO_ANNUAL_PRICE_ID")
			billing.BusinessMonthlyPriceID = os.Getenv("STRIPE_BUSINESS_MONTHLY_PRICE_ID")
			billing.BusinessAnnualPriceID = os.Getenv("STRIPE_BUSINESS_ANNUAL_PRICE_ID")
			srv.StripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")
			fmt.Fprintf(os.Stderr, "Stripe billing enabled\n")
		}

		// Google OAuth
		googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
		googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
		if googleClientID != "" && googleClientSecret != "" {
			srv.OAuthConfig = auth.GoogleOAuthConfig(
				googleClientID,
				googleClientSecret,
				srv.BaseURL+"/internal/auth/google/callback",
			)
			fmt.Fprintf(os.Stderr, "Google OAuth enabled\n")
		}

		// Embedded SPA
		distFS, err := fs.Sub(web.DistFS, "dist")
		if err == nil {
			srv.WebFS = distFS
			fmt.Fprintf(os.Stderr, "Serving embedded web UI\n")
		}

		fmt.Fprintf(os.Stderr, "Full mode active (DB + Auth + Billing)\n")
	}

	if err := srv.Start(ctx); err != nil {
		fatal("server: %v", err)
	}
}
