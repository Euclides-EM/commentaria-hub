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
	InputTokens              int64    `json:"input_tokens"`
	CachedInputTokens        int64    `json:"cached_input_tokens"`
	CacheCreationInputTokens int64    `json:"cache_creation_input_tokens"`
	CacheMetricsAvailable    bool     `json:"cache_metrics_available"`
	OutputTokens             int64    `json:"output_tokens"`
	ReasoningTokens          int64    `json:"reasoning_tokens"`
	TotalTokens              int64    `json:"total_tokens"`
	CostUSD                  *float64 `json:"cost_usd,omitempty"`
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
