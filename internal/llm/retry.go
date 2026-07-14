package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/go-logr/logr"
	"google.golang.org/genai"
)

const (
	primaryRetries  = 2
	fallbackRetries = 0
	retryDelay      = 30 * time.Second
)

var (
	retryDelayOverride time.Duration

	retryableStatusCodes = map[int]bool{
		429: true, // Rate limit
		500: true, // Internal server error
		502: true, // Bad gateway
		503: true, // Service unavailable
	}
)

func AnalyzeWithRetry(
	ctx context.Context,
	logger logr.Logger,
	primaryFn func() (*AnalysisResult, error),
	fallbackFn func() (*AnalysisResult, error),
) (*AnalysisResult, error) {
	result, exhausted, primaryErr := retryLoop(ctx, logger, "primary", primaryFn, primaryRetries)
	if primaryErr == nil {
		return result, nil
	}

	if !exhausted {
		return nil, primaryErr
	}

	logger.Info("switching to fallback model", "reason", "primary model retries exhausted")

	result, _, fallbackErr := retryLoop(ctx, logger, "fallback", fallbackFn, fallbackRetries)
	if fallbackErr != nil {
		logger.Error(errors.Join(primaryErr, fallbackErr), "LLM analysis failed after all retries on both models")
		return nil, fmt.Errorf("LLM analysis unavailable: both primary and fallback models failed after retries: %w", errors.Join(primaryErr, fallbackErr))
	}

	return result, nil
}

func retryLoop(ctx context.Context, logger logr.Logger, modelName string, fn func() (*AnalysisResult, error), maxRetries int) (*AnalysisResult, bool, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			if attempt > 0 {
				logger.Info("LLM analysis succeeded after retry", "model", modelName, "attempt", attempt)
			}
			return result, false, nil
		}

		lastErr = err

		if !isRetryable(err) {
			return nil, false, err
		}

		if attempt < maxRetries {
			backoff := retryDelay
			if retryDelayOverride > 0 {
				backoff = retryDelayOverride
			}
			logger.Info("retrying LLM analysis", "model", modelName, "attempt", attempt+1, "maxRetries", maxRetries, "backoff", backoff, "error", err.Error())

			timer := time.NewTimer(backoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, false, fmt.Errorf("retry canceled: %w (last LLM error: %v)", ctx.Err(), lastErr)
			case <-timer.C:
			}
		}
	}

	return nil, true, lastErr
}

func isRetryable(err error) bool {
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		return retryableStatusCodes[apiErr.Code]
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if errors.Is(err, context.Canceled) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, ErrNoResponseCandidates) || errors.Is(err, ErrNoContentInResponse) {
		return true
	}

	if errors.Is(err, ErrToolCallFailed) || errors.Is(err, ErrMaxIterations) {
		return false
	}

	return false
}
