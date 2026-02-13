package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/DanielTso/pixshift/internal/db"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

// VerifyWebhook verifies a Stripe webhook signature and parses the event.
func VerifyWebhook(payload []byte, signature, secret string) (*stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, secret)
	if err != nil {
		return nil, fmt.Errorf("verify webhook: %w", err)
	}
	return &event, nil
}

// ProcessEvent handles a verified Stripe webhook event by updating the database.
func ProcessEvent(ctx context.Context, database *db.DB, event *stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		return handleCheckoutCompleted(ctx, database, event)
	case "customer.subscription.deleted":
		return handleSubscriptionDeleted(ctx, database, event)
	case "customer.subscription.updated":
		return handleSubscriptionUpdated(ctx, database, event)
	case "invoice.payment_failed":
		return handlePaymentFailed(event)
	default:
		return nil
	}
}

func handleCheckoutCompleted(ctx context.Context, database *db.DB, event *stripe.Event) error {
	var session struct {
		Customer     string `json:"customer"`
		Subscription string `json:"subscription"`
	}
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("parse checkout session: %w", err)
	}

	user, err := database.GetUserByStripeCustomer(ctx, session.Customer)
	if err != nil {
		return fmt.Errorf("find user for customer %s: %w", session.Customer, err)
	}

	if err := database.UpdateUserTier(ctx, user.ID, TierPro); err != nil {
		return fmt.Errorf("set tier to pro: %w", err)
	}
	if err := database.UpdateStripeSubscription(ctx, user.ID, session.Subscription); err != nil {
		return fmt.Errorf("store subscription id: %w", err)
	}
	return nil
}

func handleSubscriptionDeleted(ctx context.Context, database *db.DB, event *stripe.Event) error {
	var sub struct {
		Customer string `json:"customer"`
	}
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("parse subscription deleted: %w", err)
	}

	user, err := database.GetUserByStripeCustomer(ctx, sub.Customer)
	if err != nil {
		return fmt.Errorf("find user for customer %s: %w", sub.Customer, err)
	}

	if err := database.UpdateUserTier(ctx, user.ID, TierFree); err != nil {
		return fmt.Errorf("set tier to free: %w", err)
	}
	if err := database.UpdateStripeSubscription(ctx, user.ID, ""); err != nil {
		return fmt.Errorf("clear subscription id: %w", err)
	}
	return nil
}

func handleSubscriptionUpdated(ctx context.Context, database *db.DB, event *stripe.Event) error {
	var sub struct {
		Customer string `json:"customer"`
		Status   string `json:"status"`
	}
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return fmt.Errorf("parse subscription updated: %w", err)
	}

	user, err := database.GetUserByStripeCustomer(ctx, sub.Customer)
	if err != nil {
		return fmt.Errorf("find user for customer %s: %w", sub.Customer, err)
	}

	tier := TierPro
	if sub.Status != "active" && sub.Status != "trialing" {
		tier = TierFree
	}
	if err := database.UpdateUserTier(ctx, user.ID, tier); err != nil {
		return fmt.Errorf("update tier: %w", err)
	}
	return nil
}

func handlePaymentFailed(event *stripe.Event) error {
	var inv struct {
		Customer string `json:"customer"`
	}
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		return fmt.Errorf("parse payment failed: %w", err)
	}
	log.Printf("payment failed for customer %s", inv.Customer)
	return nil
}
