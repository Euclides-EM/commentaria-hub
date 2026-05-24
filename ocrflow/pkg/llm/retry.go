package llm

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/openai/openai-go/v3"
)

const (
	maxRateLimitRetries = 5
	maxNetworkRetries   = 3
	baseRetryDelay      = 2 * time.Second
	maxRetryDelay       = 30 * time.Second
)

var retryAfterMessagePattern = regexp.MustCompile(`Please try again in ([0-9hms.]+)`)

func executeWithRetries(ctx context.Context, model string, call func() error) (uint, error) {
	return executeWithRetriesForAttempts(ctx, model, maxRateLimitRetries+1, call)
}

func executeWithRetriesForAttempts(ctx context.Context, model string, attemptsLimit uint, call func() error) (uint, error) {
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
			log.Printf("debug: llm exec retry model=%s attempt=%d retry_in=%s", model, n+1, delay)
		}),
	)
	return attempts, err
}

func retryDelay(err error, attempt uint) (time.Duration, bool) {
	if statusCode, resp, ok := retryableHTTPResponse(err); ok {
		if !isRetryableStatusCode(statusCode) {
			return 0, false
		}
		if delay := retryDelayFromHeaders(resp); delay > 0 {
			return minDuration(delay, maxRetryDelay), true
		}
		if delay := retryDelayFromMessage(err.Error()); delay > 0 {
			return minDuration(delay, maxRetryDelay), true
		}
		return retryBackoff(attempt), true
	}

	if !isRetryableNetworkError(err) {
		return 0, false
	}

	return retryBackoff(attempt), true
}

type httpStatusError struct {
	statusCode int
	response   *http.Response
	message    string
}

func (e *httpStatusError) Error() string {
	return e.message
}

func newHTTPStatusError(statusCode int, response *http.Response, message string) error {
	return &httpStatusError{
		statusCode: statusCode,
		response:   response,
		message:    message,
	}
}

func retryableHTTPResponse(err error) (int, *http.Response, bool) {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode, apiErr.Response, true
	}

	var httpErr *httpStatusError
	if errors.As(err, &httpErr) {
		return httpErr.statusCode, httpErr.response, true
	}

	return 0, nil, false
}

func isRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || (statusCode >= http.StatusInternalServerError && statusCode <= http.StatusNetworkAuthenticationRequired)
}

func retryBackoff(attempt uint) time.Duration {
	backoff := baseRetryDelay * time.Duration(1<<(attempt-1))
	return minDuration(backoff, maxRetryDelay)
}

func isRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	return false
}

func retryDelayFromHeaders(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}

	for _, key := range []string{"retry-after-ms", "Retry-After-Ms"} {
		if value := strings.TrimSpace(resp.Header.Get(key)); value != "" {
			ms, err := strconv.Atoi(value)
			if err == nil && ms > 0 {
				return time.Duration(ms) * time.Millisecond
			}
		}
	}

	if value := strings.TrimSpace(resp.Header.Get("Retry-After")); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
			return time.Duration(seconds) * time.Second
		}
		if at, err := http.ParseTime(value); err == nil {
			return time.Until(at)
		}
	}

	for _, key := range []string{"x-ratelimit-reset-requests", "x-ratelimit-reset-tokens"} {
		if value := strings.TrimSpace(resp.Header.Get(key)); value != "" {
			if delay, err := time.ParseDuration(value); err == nil && delay > 0 {
				return delay
			}
		}
	}

	return 0
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

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
