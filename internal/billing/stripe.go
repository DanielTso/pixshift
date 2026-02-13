package billing

import (
	stripe "github.com/stripe/stripe-go/v82"
)

// Tier constants for subscription plans.
const (
	TierFree     = "free"
	TierPro      = "pro"
	TierBusiness = "business"
)

// Stripe price IDs for each plan and billing interval.
var (
	ProMonthlyPriceID      string
	ProAnnualPriceID       string
	BusinessMonthlyPriceID string
	BusinessAnnualPriceID  string
)

// TierForPriceID maps a Stripe price ID to a tier constant.
// Defaults to TierPro for backward compatibility with unknown price IDs.
func TierForPriceID(priceID string) string {
	switch priceID {
	case BusinessMonthlyPriceID, BusinessAnnualPriceID:
		if priceID != "" {
			return TierBusiness
		}
	}
	return TierPro
}

// Init sets the Stripe API key.
func Init(apiKey string) {
	stripe.Key = apiKey
}
