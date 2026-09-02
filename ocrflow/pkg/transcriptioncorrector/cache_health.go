package transcriptioncorrector

import (
	"log"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
)

type cacheHealthTracker struct {
	provider              string
	model                 string
	requests              int
	readOpportunities     int
	readRequests          int
	missesAfterWarmup     int
	consecutiveMisses     int
	cachedTokens          int64
	metricsAvailable      bool
	metricsUnavailableLog bool
}

func newCacheHealthTracker(provider, model string) *cacheHealthTracker {
	return &cacheHealthTracker{
		provider: strings.ToLower(strings.TrimSpace(provider)),
		model:    strings.TrimSpace(model),
	}
}

func (t *cacheHealthTracker) Observe(usage llm.Usage, logger *log.Logger) {
	t.requests++
	if !usage.CacheMetricsAvailable {
		if t.requests >= 2 && !t.metricsUnavailableLog {
			logger.Printf("warning: prompt cache cannot be verified provider=%s model=%s cached_tokens metric unavailable", t.provider, t.model)
			t.metricsUnavailableLog = true
		}
		return
	}
	t.metricsAvailable = true
	t.cachedTokens += usage.CachedInputTokens
	if t.requests == 1 {
		return
	}

	t.readOpportunities++
	if usage.CachedInputTokens > 0 {
		t.readRequests++
		t.consecutiveMisses = 0
		return
	}

	t.missesAfterWarmup++
	t.consecutiveMisses++
	if t.consecutiveMisses == 1 || t.consecutiveMisses%25 == 0 {
		logger.Printf(
			"warning: prompt cache miss after warmup provider=%s model=%s request=%d consecutive_misses=%d cached_tokens=0; verify stable-prefix ordering, cache eligibility, and TTL",
			t.provider, t.model, t.requests, t.consecutiveMisses,
		)
	}
}

func (t *cacheHealthTracker) LogSummary(logger *log.Logger) {
	logger.Printf(
		"cache summary provider=%s model=%s requests=%d metrics_available=%t read_opportunities=%d cache_read_requests=%d misses_after_warmup=%d cached_tokens=%d",
		t.provider, t.model, t.requests, t.metricsAvailable, t.readOpportunities,
		t.readRequests, t.missesAfterWarmup, t.cachedTokens,
	)
}
