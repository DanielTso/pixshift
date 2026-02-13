package billing

import (
	stripe "github.com/stripe/stripe-go/v82"
)

// Tier constants for subscription plans.
const (
	TierFree = "free"
	TierPro  = "pro"
)

// ProPriceID should be set from environment (STRIPE_PRICE_ID).
var ProPriceID string

// Init sets the Stripe API key.
func Init(apiKey string) {
	stripe.Key = apiKey
}
