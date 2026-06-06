package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
)

func RetryableHTTPResponse(err error) (int, time.Duration, bool) {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode, retryDelayFromHeaders(apiErr.Response), true
	}

	var httpErr *StatusError
	if errors.As(err, &httpErr) {
		return httpErr.statusCode, retryDelayFromHeaders(httpErr.response), true
	}

	return 0, 0, false
}

func IsRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || (statusCode >= http.StatusInternalServerError && statusCode <= http.StatusNetworkAuthenticationRequired)
}

func IsRetryableNetworkError(err error) bool {
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

	if resp.Body != nil {
		defer resp.Body.Close()
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
