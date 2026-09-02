package llm

import "testing"

func TestUsageAddAccumulatesTokensAndReportedCost(t *testing.T) {
	firstCost := 0.01
	secondCost := 0.02
	usage := Usage{}
	usage.Add(Usage{
		InputTokens: 10, CachedInputTokens: 2, CacheCreationInputTokens: 3,
		CacheMetricsAvailable: true, OutputTokens: 4, ReasoningTokens: 1, TotalTokens: 19, CostUSD: &firstCost,
	})
	usage.Add(Usage{
		InputTokens: 20, CachedInputTokens: 4, CacheCreationInputTokens: 6,
		OutputTokens: 8, ReasoningTokens: 2, TotalTokens: 38, CostUSD: &secondCost,
	})

	if usage.InputTokens != 30 || usage.CachedInputTokens != 6 || usage.CacheCreationInputTokens != 9 ||
		usage.OutputTokens != 12 || usage.ReasoningTokens != 3 || usage.TotalTokens != 57 {
		t.Fatalf("usage = %#v, want accumulated token counts", usage)
	}
	if !usage.CacheMetricsAvailable {
		t.Fatal("cache metrics availability was not preserved")
	}
	if usage.CostUSD == nil || *usage.CostUSD != 0.03 {
		t.Fatalf("cost = %v, want 0.03", usage.CostUSD)
	}
}
