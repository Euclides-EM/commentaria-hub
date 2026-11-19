package transcriptioncorrector

import (
	"log"
	"strings"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/stretchr/testify/require"
)

func TestCacheHealthWarnsOnMissAfterWarmup(t *testing.T) {
	var logs strings.Builder
	logger := log.New(&logs, "", 0)
	tracker := newCacheHealthTracker(llm.ProviderOpenAI, "gpt-5.6")

	tracker.Observe(llm.Usage{CacheMetricsAvailable: true}, logger)
	tracker.Observe(llm.Usage{CacheMetricsAvailable: true}, logger)
	tracker.Observe(llm.Usage{CacheMetricsAvailable: true, CachedInputTokens: 1200}, logger)
	tracker.LogSummary(logger)

	require.Contains(t, logs.String(), "warning: prompt cache miss after warmup")
	require.Contains(t, logs.String(), "cache_read_requests=1")
	require.Contains(t, logs.String(), "misses_after_warmup=1")
	require.Contains(t, logs.String(), "cached_tokens=1200")
}

func TestCacheHealthWarnsOnceWhenMetricsUnavailable(t *testing.T) {
	var logs strings.Builder
	logger := log.New(&logs, "", 0)
	tracker := newCacheHealthTracker(llm.ProviderOllama, "vision")

	tracker.Observe(llm.Usage{}, logger)
	tracker.Observe(llm.Usage{}, logger)
	tracker.Observe(llm.Usage{}, logger)

	require.Equal(t, 1, strings.Count(logs.String(), "prompt cache cannot be verified"))
}
