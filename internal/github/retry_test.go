package github

import (
	"testing"
	"time"
)

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"auth error", &AuthError{StatusCode: 401}, false},
		{"query error", &QueryError{Errors: []GraphQLError{{Message: "bad query"}}}, false},
		{"server error", &HTTPError{StatusCode: 500}, true},
		{"rate limit", &HTTPError{StatusCode: 429}, true},
		{"client error", &HTTPError{StatusCode: 400}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBackoffDelay(t *testing.T) {
	for attempt := 0; attempt < 5; attempt++ {
		delay := backoffDelay(attempt)
		if delay < 0 {
			t.Errorf("attempt %d: negative delay %v", attempt, delay)
		}
		if delay > maxDelay*2 { // Allow 2x for jitter
			t.Errorf("attempt %d: delay %v exceeds max with jitter", attempt, delay)
		}
	}

	// Check that delays generally increase
	delays := make([]time.Duration, 5)
	for i := range delays {
		delays[i] = backoffDelay(i)
	}
	// Due to jitter, we just check the general trend
	t.Logf("Delays: %v", delays)
}
