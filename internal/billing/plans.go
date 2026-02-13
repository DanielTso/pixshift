package billing

// TierLimits defines the usage limits for a subscription tier.
type TierLimits struct {
	MaxConversionsPerDay   int // 0 = unlimited
	MaxFileSizeMB          int
	RateLimitPerMin        int
	MaxAPIKeys             int
	MaxAPIRequestsPerMonth int // 0 = unlimited
	MaxBatchSize           int
}

// GetLimits returns the usage limits for the given tier.
func GetLimits(tier string) TierLimits {
	switch tier {
	case TierBusiness:
		return TierLimits{
			MaxConversionsPerDay:   0,
			MaxFileSizeMB:          500,
			RateLimitPerMin:        120,
			MaxAPIKeys:             20,
			MaxAPIRequestsPerMonth: 50000,
			MaxBatchSize:           100,
		}
	case TierPro:
		return TierLimits{
			MaxConversionsPerDay:   500,
			MaxFileSizeMB:          100,
			RateLimitPerMin:        30,
			MaxAPIKeys:             5,
			MaxAPIRequestsPerMonth: 5000,
			MaxBatchSize:           20,
		}
	default:
		return TierLimits{
			MaxConversionsPerDay:   20,
			MaxFileSizeMB:          10,
			RateLimitPerMin:        10,
			MaxAPIKeys:             1,
			MaxAPIRequestsPerMonth: 100,
			MaxBatchSize:           1,
		}
	}
}
