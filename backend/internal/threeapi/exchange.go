package threeapi

// TAI cost calculation for API calls.
// Anchor: 1 TAI ≈ $0.001 compute cost (1 basic API call).
// Model tiers determine TAI cost per 1000 tokens.

// ModelTier defines pricing per model category.
type ModelTier struct {
	Name       string
	TAIPer1KTokens float64 // TAI cost per 1000 tokens
}

var modelTiers = map[string]ModelTier{
	// Basic tier: 1 TAI per ~1000 tokens
	"gpt-4o-mini":       {Name: "basic", TAIPer1KTokens: 1.0},
	"qwen-turbo":        {Name: "basic", TAIPer1KTokens: 0.8},
	"qwen-plus":         {Name: "basic", TAIPer1KTokens: 1.2},
	"gemini-flash":      {Name: "basic", TAIPer1KTokens: 0.9},

	// Mid tier: ~5 TAI per 1000 tokens
	"gpt-4o":            {Name: "mid", TAIPer1KTokens: 5.0},
	"claude-sonnet-4-20250514": {Name: "mid", TAIPer1KTokens: 5.0},
	"qwen-max":          {Name: "mid", TAIPer1KTokens: 4.5},
	"gemini-pro":        {Name: "mid", TAIPer1KTokens: 4.0},

	// Premium tier: ~20 TAI per 1000 tokens
	"claude-opus-4-20250514":   {Name: "premium", TAIPer1KTokens: 20.0},
	"o3":                {Name: "premium", TAIPer1KTokens: 18.0},
	"gemini-ultra":      {Name: "premium", TAIPer1KTokens: 16.0},

	// Image generation: flat rate per image
	"gpt-image-2":       {Name: "image", TAIPer1KTokens: 0}, // handled separately
	"dall-e-3":          {Name: "image", TAIPer1KTokens: 0},
}

// CalculateTAICost computes TAI cost for a given model and token count.
func CalculateTAICost(model string, tokensUsed int64) float64 {
	tier, ok := modelTiers[model]
	if !ok {
		// Unknown model: default to mid tier
		tier = ModelTier{Name: "mid", TAIPer1KTokens: 5.0}
	}

	if tier.Name == "image" {
		// Image gen: flat 10 TAI per image (token count irrelevant)
		return 10.0
	}

	cost := float64(tokensUsed) / 1000.0 * tier.TAIPer1KTokens

	// Minimum cost: 0.1 TAI per call (prevent zero-cost abuse)
	if cost < 0.1 {
		cost = 0.1
	}

	return cost
}

// EstimateTaskCost estimates TAI cost for a bounty task before execution.
// Used by canExecute to check if pet can afford the task.
func EstimateTaskCost(model string, expectedTokens int64) float64 {
	return CalculateTAICost(model, expectedTokens)
}

// ModelForDifficulty returns the appropriate model for a task difficulty.
func ModelForDifficulty(difficulty string) string {
	switch difficulty {
	case "D":
		return "gpt-4o-mini"
	case "C":
		return "qwen-turbo"
	case "B":
		return "gpt-4o"
	case "A":
		return "claude-sonnet-4-20250514"
	case "S":
		return "claude-opus-4-20250514"
	default:
		return "gpt-4o-mini"
	}
}
