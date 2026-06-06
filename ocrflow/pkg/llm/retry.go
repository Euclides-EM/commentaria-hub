package llm

import (
	"context"
	"log"
	"regexp"
	"time"

	phttp "github.com/MiaMish/elements-dh/ocrflow/pkg/http"
	"github.com/avast/retry-go"
	"github.com/samber/lo"
)

const (
	maxRateLimitRetries = 5
	maxNetworkRetries   = 3
	baseRetryDelay      = 2 * time.Second
	maxRetryDelay       = 30 * time.Second
)

var retryAfterMessagePattern = regexp.MustCompile(`Please try again in ([0-9hms.]+)`)

func executeWithRetries(ctx context.Context, call func() error) (uint, error) {
	return executeWithRetriesForAttempts(ctx, maxRateLimitRetries+1, call)
}

func executeWithRetriesForAttempts(ctx context.Context, attemptsLimit uint, call func() error) (uint, error) {
	var attempts uint
	err := retry.Do(
		func() error {
			attempts++
			return call()
		},
		retry.Context(ctx),
		retry.Attempts(attemptsLimit),
		retry.LastErrorOnly(true),
		retry.DelayType(func(n uint, err error, _ *retry.Config) time.Duration {
			delay, _ := retryDelay(err, n+1)
			if delay <= 0 {
				return baseRetryDelay
			}
			return delay
		}),
		retry.RetryIf(func(err error) bool {
			_, shouldRetry := retryDelay(err, attempts)
			return shouldRetry
		}),
		retry.OnRetry(func(n uint, err error) {
			delay, _ := retryDelay(err, n+1)
			log.Printf("debug: llm exec retry attempt=%d retry_in=%s: err %v", n+1, delay, err)
		}),
	)
	return attempts, err
}

func retryDelay(err error, attempt uint) (time.Duration, bool) {
	if statusCode, delay, ok := phttp.RetryableHTTPResponse(err); ok {

		if !phttp.IsRetryableStatusCode(statusCode) {
			return 0, false
		}
		if delay > 0 {
			return lo.Min([]time.Duration{delay, maxRetryDelay}), true
		}
		if delay := retryDelayFromMessage(err.Error()); delay > 0 {
			return lo.Min([]time.Duration{delay, maxRetryDelay}), true
		}
		return retryBackoff(attempt), true
	}

	if !phttp.IsRetryableNetworkError(err) {
		return 0, false
	}

	return retryBackoff(attempt), true
}

func retryBackoff(attempt uint) time.Duration {
	backoff := baseRetryDelay * time.Duration(1<<(attempt-1))
	return lo.Min([]time.Duration{backoff, maxRetryDelay})
}

func retryDelayFromMessage(message string) time.Duration {
	match := retryAfterMessagePattern.FindStringSubmatch(message)
	if len(match) != 2 {
		return 0
	}
	delay, err := time.ParseDuration(match[1])
	if err != nil || delay <= 0 {
		return 0
	}
	return delay
}
