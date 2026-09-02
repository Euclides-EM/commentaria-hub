package llm

// Result contains the generated text and the provider-reported usage for one
// successful LLM request. CostUSD is nil when the provider does not report a
// request-level cost.
type Result struct {
	Text  string
	Usage Usage
}

// Prompt separates a stable, reusable instruction prefix from request-specific
// input. Providers may place Static in a cacheable system/developer block while
// keeping Dynamic and the attachment after the cache boundary.
type Prompt struct {
	Static   string
	Dynamic  string
	CacheKey string
}

// Usage preserves the token counters reported by each provider. OpenAI's
// InputTokens includes cached input, while Claude Code reports cache reads and
// cache creation separately.
type Usage struct {
	InputTokens              int64
	CachedInputTokens        int64
	CacheCreationInputTokens int64
	CacheMetricsAvailable    bool
	OutputTokens             int64
	ReasoningTokens          int64
	TotalTokens              int64
	CostUSD                  *float64
}

// Add accumulates token counts and any provider-reported cost.
func (u *Usage) Add(other Usage) {
	u.InputTokens += other.InputTokens
	u.CachedInputTokens += other.CachedInputTokens
	u.CacheCreationInputTokens += other.CacheCreationInputTokens
	u.CacheMetricsAvailable = u.CacheMetricsAvailable || other.CacheMetricsAvailable
	u.OutputTokens += other.OutputTokens
	u.ReasoningTokens += other.ReasoningTokens
	u.TotalTokens += other.TotalTokens
	if other.CostUSD != nil {
		if u.CostUSD == nil {
			zero := float64(0)
			u.CostUSD = &zero
		}
		*u.CostUSD += *other.CostUSD
	}
}
