package threeapi

// ExchangeRate defines the TAI → 3api credit conversion.
// 1 TAI ≈ 1 basic API call on 3api.
// The actual credit amount depends on the model tier.
type ExchangeRate struct {
	// BasicModel: GPT-4o-mini, Qwen-turbo, etc. 1 TAI = 1 call
	BasicModel float64
	// MidModel: GPT-4o, Claude Sonnet, Qwen-max. 1 TAI = 0.2 calls (5 TAI per call)
	MidModel float64
	// PremiumModel: Claude Opus, o3. 1 TAI = 0.05 calls (20 TAI per call)
	PremiumModel float64
	// ImageGen: GPT Image 2, DALL-E. 1 TAI = 0.1 calls (10 TAI per image)
	ImageGen float64
}

// DefaultExchangeRate returns the standard pricing tier.
// These map to 3api's actual per-call costs in USD, converted to TAI.
func DefaultExchangeRate() ExchangeRate {
	return ExchangeRate{
		BasicModel:   1.0,  // 1 TAI per call
		MidModel:     5.0,  // 5 TAI per call
		PremiumModel: 20.0, // 20 TAI per call
		ImageGen:     10.0, // 10 TAI per image
	}
}

// CreditForCall calculates how much 3api balance (in USD equivalent)
// to credit for a given TAI spend and model tier.
// The 3api balance is denominated in USD internally.
func CreditForCall(taiAmount float64, tier string) float64 {
	rate := DefaultExchangeRate()

	// USD value per TAI (anchor: 1 TAI ≈ $0.001 compute cost)
	const usdPerTAI = 0.001

	switch tier {
	case "basic":
		return taiAmount * usdPerTAI
	case "mid":
		return taiAmount * usdPerTAI
	case "premium":
		return taiAmount * usdPerTAI
	case "image":
		return taiAmount * usdPerTAI
	default:
		return taiAmount * usdPerTAI
	}
	_ = rate
}

// TAICostForTask estimates TAI cost for an agent task based on
// expected number of API calls and model tiers used.
func TAICostForTask(expectedCalls int, primaryTier string) float64 {
	rate := DefaultExchangeRate()

	var costPerCall float64
	switch primaryTier {
	case "basic":
		costPerCall = rate.BasicModel
	case "mid":
		costPerCall = rate.MidModel
	case "premium":
		costPerCall = rate.PremiumModel
	case "image":
		costPerCall = rate.ImageGen
	default:
		costPerCall = rate.BasicModel
	}

	return float64(expectedCalls) * costPerCall
}
