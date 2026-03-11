package llm

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
)

func TestRetryDelayFromHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers http.Header
		want    time.Duration
		min     time.Duration
		max     time.Duration
	}{
		{
			name: "retry after ms",
			headers: http.Header{
				"Retry-After-Ms": []string{"1500"},
			},
			want: 1500 * time.Millisecond,
		},
		{
			name: "retry after seconds",
			headers: http.Header{
				"Retry-After": []string{"3"},
			},
			want: 3 * time.Second,
		},
		{
			name: "ratelimit reset duration",
			headers: http.Header{
				"X-Ratelimit-Reset-Requests": []string{"4s"},
			},
			want: 4 * time.Second,
		},
		{
			name: "retry after http date",
			headers: http.Header{
				"Retry-After": []string{time.Now().Add(2 * time.Second).UTC().Format(http.TimeFormat)},
			},
			min: 1 * time.Second,
			max: 3 * time.Second,
		},
		{
			name:    "missing headers",
			headers: http.Header{},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{Header: tt.headers}
			got := retryDelayFromHeaders(resp)
			if tt.min != 0 || tt.max != 0 {
				if got < tt.min || got > tt.max {
					t.Fatalf("retryDelayFromHeaders() = %s, want between %s and %s", got, tt.min, tt.max)
				}
				return
			}
			if got != tt.want {
				t.Fatalf("retryDelayFromHeaders() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRetryDelayFromMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    time.Duration
	}{
		{
			name:    "parses duration",
			message: `POST "https://api.openai.com/v1/responses": 429 {"message":"Please try again in 12.5s"}`,
			want:    12500 * time.Millisecond,
		},
		{
			name:    "ignores unrelated message",
			message: "plain error",
			want:    0,
		},
		{
			name:    "ignores invalid duration",
			message: `{"message":"Please try again in soon"}`,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryDelayFromMessage(tt.message)
			if got != tt.want {
				t.Fatalf("retryDelayFromMessage() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		attempt uint
		want    time.Duration
		retry   bool
	}{
		{
			name:    "non openai error",
			err:     assertErr("boom"),
			attempt: 1,
			want:    0,
			retry:   false,
		},
		{
			name:    "non rate limit openai error",
			err:     newOpenAIError(http.StatusBadRequest, http.Header{}, `{"message":"bad request","type":"invalid_request_error","param":null,"code":null}`),
			attempt: 1,
			want:    0,
			retry:   false,
		},
		{
			name:    "server error retries with exponential fallback",
			err:     newOpenAIError(http.StatusInternalServerError, http.Header{}, `{"message":"server error","type":"server_error","param":null,"code":null}`),
			attempt: 2,
			want:    4 * time.Second,
			retry:   true,
		},
		{
			name:    "server error uses retry after header",
			err:     newOpenAIError(http.StatusBadGateway, http.Header{"Retry-After": []string{"6"}}, `{"message":"bad gateway","type":"server_error","param":null,"code":null}`),
			attempt: 1,
			want:    6 * time.Second,
			retry:   true,
		},
		{
			name:    "uses retry after header",
			err:     newOpenAIError(http.StatusTooManyRequests, http.Header{"Retry-After": []string{"4"}}, `{"message":"rate limited","type":"rate_limit_error","param":null,"code":null}`),
			attempt: 1,
			want:    4 * time.Second,
			retry:   true,
		},
		{
			name:    "uses message fallback",
			err:     newOpenAIError(http.StatusTooManyRequests, http.Header{}, `{"message":"Please try again in 5s","type":"rate_limit_error","param":null,"code":null}`),
			attempt: 1,
			want:    5 * time.Second,
			retry:   true,
		},
		{
			name:    "uses exponential fallback",
			err:     newOpenAIError(http.StatusTooManyRequests, http.Header{}, `{"message":"rate limited","type":"rate_limit_error","param":null,"code":null}`),
			attempt: 3,
			want:    8 * time.Second,
			retry:   true,
		},
		{
			name:    "caps large delay",
			err:     newOpenAIError(http.StatusTooManyRequests, http.Header{"Retry-After": []string{"120"}}, `{"message":"rate limited","type":"rate_limit_error","param":null,"code":null}`),
			attempt: 1,
			want:    maxRetryDelay,
			retry:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, retry := retryDelay(tt.err, tt.attempt)
			if retry != tt.retry {
				t.Fatalf("retryDelay() retry = %t, want %t", retry, tt.retry)
			}
			if got != tt.want {
				t.Fatalf("retryDelay() delay = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestMinDuration(t *testing.T) {
	if got := minDuration(2*time.Second, 5*time.Second); got != 2*time.Second {
		t.Fatalf("minDuration() = %s, want %s", got, 2*time.Second)
	}
	if got := minDuration(6*time.Second, 5*time.Second); got != 5*time.Second {
		t.Fatalf("minDuration() = %s, want %s", got, 5*time.Second)
	}
}

func newOpenAIError(status int, headers http.Header, raw string) error {
	reqURL, err := url.Parse("https://api.openai.com/v1/responses")
	if err != nil {
		panic(err)
	}
	apiErr := &openai.Error{
		StatusCode: status,
		Request: &http.Request{
			Method: http.MethodPost,
			URL:    reqURL,
		},
		Response: &http.Response{
			StatusCode: status,
			Header:     headers,
		},
	}
	if err := apiErr.UnmarshalJSON([]byte(raw)); err != nil {
		panic(err)
	}
	return apiErr
}

func assertErr(message string) error {
	return &staticError{message: message}
}

type staticError struct {
	message string
}

func (e *staticError) Error() string {
	return strings.TrimSpace(e.message)
}
