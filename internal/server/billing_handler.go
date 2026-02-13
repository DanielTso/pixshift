package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/billing"
)

// handleCheckout handles POST /internal/billing/checkout.
func (s *Server) handleCheckout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	// Create Stripe customer if needed
	customerID := ""
	if user.StripeCustomerID.Valid {
		customerID = user.StripeCustomerID.String
	} else {
		var err error
		customerID, err = billing.CreateCustomer(user.Email, user.Name)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "STRIPE_ERROR", "failed to create customer")
			return
		}
		if err := s.DB.UpdateStripeCustomer(r.Context(), user.ID, customerID); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			return
		}
	}

	successURL := s.BaseURL + "/dashboard?upgraded=1"
	cancelURL := s.BaseURL + "/pricing"

	url, err := billing.CreateCheckoutSession(customerID, billing.ProPriceID, successURL, cancelURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STRIPE_ERROR", "failed to create checkout session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// handlePortal handles POST /internal/billing/portal.
func (s *Server) handlePortal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	if !user.StripeCustomerID.Valid {
		writeError(w, http.StatusBadRequest, "NO_SUBSCRIPTION", "no active subscription")
		return
	}

	returnURL := s.BaseURL + "/dashboard"

	url, err := billing.CreatePortalSession(user.StripeCustomerID.String, returnURL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "STRIPE_ERROR", "failed to create portal session")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// handleStripeWebhook handles POST /api/webhooks/stripe.
func (s *Server) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	// Limit body to 65536 bytes to prevent abuse
	body, err := io.ReadAll(io.LimitReader(r.Body, 65536))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "failed to read request body")
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := billing.VerifyWebhook(body, signature, s.StripeWebhookSecret)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_SIGNATURE", "invalid webhook signature")
		return
	}

	if err := billing.ProcessEvent(r.Context(), s.DB, event); err != nil {
		writeError(w, http.StatusInternalServerError, "WEBHOOK_ERROR", "failed to process webhook event")
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
