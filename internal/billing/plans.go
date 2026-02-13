package billing

// TierLimits defines the usage limits for a subscription tier.
type TierLimits struct {
	MaxConversionsPerDay int // 0 = unlimited
	MaxFileSizeMB        int
	RateLimitPerMin      int
	MaxAPIKeys           int
}

// GetLimits returns the usage limits for the given tier.
func GetLimits(tier string) TierLimits {
	switch tier {
	case TierPro:
		return TierLimits{
			MaxConversionsPerDay: 0,
			MaxFileSizeMB:        100,
			RateLimitPerMin:      60,
			MaxAPIKeys:           10,
		}
	default:
		return TierLimits{
			MaxConversionsPerDay: 20,
			MaxFileSizeMB:        10,
			RateLimitPerMin:      10,
			MaxAPIKeys:           1,
		}
	}
}
