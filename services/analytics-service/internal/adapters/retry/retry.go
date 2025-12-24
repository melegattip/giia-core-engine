// Package retry provides exponential backoff retry logic.
package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

// Config holds retry configuration.
type Config struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	JitterFactor   float64
}

// DefaultConfig returns a default retry configuration.
func DefaultConfig() Config {
	return Config{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     10 * time.Second,
		Multiplier:     2.0,
		JitterFactor:   0.2,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc func() error

// IsRetryable determines if an error is retryable.
type IsRetryable func(error) bool

// DefaultIsRetryable returns true for all errors except context cancellation.
func DefaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return true
}

// Do executes the function with retry logic.
func Do(ctx context.Context, config Config, fn RetryableFunc, isRetryable IsRetryable) error {
	if isRetryable == nil {
		isRetryable = DefaultIsRetryable
	}

	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if !isRetryable(err) {
			return err
		}

		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff with jitter
		jitter := backoff.Seconds() * config.JitterFactor * (rand.Float64()*2 - 1)
		sleepTime := time.Duration(backoff.Seconds()+jitter) * time.Second
		if sleepTime < 0 {
			sleepTime = backoff
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepTime):
		}

		// Increase backoff for next attempt
		backoff = time.Duration(float64(backoff) * config.Multiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return lastErr
}

// DoWithResult executes a function that returns a result with retry logic.
func DoWithResult[T any](ctx context.Context, config Config, fn func() (T, error), isRetryable IsRetryable) (T, error) {
	var result T
	var lastErr error
	backoff := config.InitialBackoff

	if isRetryable == nil {
		isRetryable = DefaultIsRetryable
	}

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !isRetryable(err) {
			return result, err
		}

		if attempt == config.MaxRetries {
			break
		}

		// Calculate backoff with jitter
		jitter := backoff.Seconds() * config.JitterFactor * (rand.Float64()*2 - 1)
		sleepTime := time.Duration(math.Max(0, backoff.Seconds()+jitter) * float64(time.Second))

		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(sleepTime):
		}

		// Increase backoff for next attempt
		backoff = time.Duration(float64(backoff) * config.Multiplier)
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}
	}

	return result, lastErr
}
