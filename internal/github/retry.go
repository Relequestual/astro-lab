package github

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"
)

const (
	maxRetries = 4
	baseDelay  = 500 * time.Millisecond
	maxDelay   = 8 * time.Second
)

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	// Don't retry auth errors
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return false
	}
	// Don't retry query errors (malformed queries)
	var queryErr *QueryError
	if errors.As(err, &queryErr) {
		return false
	}
	// Retry HTTP errors (5xx, rate limits)
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 500 || httpErr.StatusCode == 429
	}
	// Retry other transient errors (network issues)
	return true
}

// DoWithRetry executes a GraphQL request with bounded retries
func (c *Client) DoWithRetry(ctx context.Context, req GraphQLRequest) (*GraphQLResponse, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.Do(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !IsRetryable(err) {
			return nil, err
		}
		if attempt < maxRetries {
			delay := backoffDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return nil, lastErr
}

func backoffDelay(attempt int) time.Duration {
	delay := float64(baseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(maxDelay) {
		delay = float64(maxDelay)
	}
	// Add jitter: 0.5x to 1.5x
	jitter := 0.5 + rand.Float64()
	return time.Duration(delay * jitter)
}
