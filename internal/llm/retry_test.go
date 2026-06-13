package llm

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genai"
)

type mockNetError struct {
	timeout bool
}

func (e *mockNetError) Error() string   { return "network error" }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return true }

func TestAnalyzeWithRetry_SuccessFirstAttempt(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		return &AnalysisResult{Content: "success from primary"}, nil
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return &AnalysisResult{Content: "success from fallback"}, nil
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.NoError(t, err)
	assert.Equal(t, "success from primary", result.Content)
	assert.Equal(t, 1, primaryCallCount)
	assert.Equal(t, 0, fallbackCallCount)
}

func TestAnalyzeWithRetry_SuccessAfterTransientFailure(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		if primaryCallCount == 1 {
			return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 429, Message: "rate limited"})
		}
		return &AnalysisResult{Content: "success after retry"}, nil
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return &AnalysisResult{Content: "fallback"}, nil
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.NoError(t, err)
	assert.Equal(t, "success after retry", result.Content)
	assert.Equal(t, 2, primaryCallCount)
	assert.Equal(t, 0, fallbackCallCount)
}

func TestAnalyzeWithRetry_PrimaryExhaustedFallbackSucceeds(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 503, Message: "service unavailable"})
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return &AnalysisResult{Content: "success from fallback"}, nil
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.NoError(t, err)
	assert.Equal(t, "success from fallback", result.Content)
	assert.Equal(t, 3, primaryCallCount)
	assert.Equal(t, 1, fallbackCallCount)
}

func TestAnalyzeWithRetry_BothModelsExhausted(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 500, Message: "internal server error"})
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 500, Message: "internal server error"})
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "LLM analysis unavailable")
	assert.Equal(t, 3, primaryCallCount)
	assert.Equal(t, 1, fallbackCallCount)
}

func TestAnalyzeWithRetry_NonRetryableErrorNoPrimaryRetry(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 400, Message: "bad request"})
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return &AnalysisResult{Content: "fallback"}, nil
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 1, primaryCallCount)
	assert.Equal(t, 0, fallbackCallCount)
}

func TestAnalyzeWithRetry_NonRetryableErrorOnFallback(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 503, Message: "service unavailable"})
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 401, Message: "unauthorized"})
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, 3, primaryCallCount)
	assert.Equal(t, 1, fallbackCallCount)
	assert.ErrorContains(t, err, "503")
	assert.ErrorContains(t, err, "401")
}

func TestAnalyzeWithRetry_RetryableStatusCodes(t *testing.T) {
	testCases := []struct {
		name string
		code int
	}{
		{"rate limit", 429},
		{"internal server error", 500},
		{"bad gateway", 502},
		{"service unavailable", 503},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryDelayOverride = time.Millisecond
			t.Cleanup(func() { retryDelayOverride = 0 })

			primaryCallCount := 0
			fallbackCallCount := 0

			primaryFn := func() (*AnalysisResult, error) {
				primaryCallCount++
				if primaryCallCount == 1 {
					return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: tc.code, Message: "error"})
				}
				return &AnalysisResult{Content: "success"}, nil
			}

			fallbackFn := func() (*AnalysisResult, error) {
				fallbackCallCount++
				return &AnalysisResult{Content: "fallback"}, nil
			}

			result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

			require.NoError(t, err)
			assert.Equal(t, "success", result.Content)
			assert.Equal(t, 2, primaryCallCount, "Expected retry for status code %d", tc.code)
			assert.Equal(t, 0, fallbackCallCount)
		})
	}
}

func TestAnalyzeWithRetry_NonRetryableStatusCodes(t *testing.T) {
	testCases := []struct {
		name string
		code int
	}{
		{"bad request", 400},
		{"unauthorized", 401},
		{"forbidden", 403},
		{"not found", 404},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryDelayOverride = time.Millisecond
			t.Cleanup(func() { retryDelayOverride = 0 })

			primaryCallCount := 0
			fallbackCallCount := 0

			primaryFn := func() (*AnalysisResult, error) {
				primaryCallCount++
				return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: tc.code, Message: "error"})
			}

			fallbackFn := func() (*AnalysisResult, error) {
				fallbackCallCount++
				return &AnalysisResult{Content: "fallback"}, nil
			}

			result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

			require.Error(t, err)
			assert.Nil(t, result)
			assert.Equal(t, 1, primaryCallCount, "Expected no retry for status code %d", tc.code)
			assert.Equal(t, 0, fallbackCallCount)
		})
	}
}

func TestAnalyzeWithRetry_ContextCancellation(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	primaryCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		if primaryCallCount == 2 {
			cancel()
		}
		return nil, fmt.Errorf("gemini API error: %w", genai.APIError{Code: 503, Message: "service unavailable"})
	}

	fallbackFn := func() (*AnalysisResult, error) {
		t.Fatal("fallback should not be called when context is canceled")
		return nil, fmt.Errorf("fallback should not be called when context is canceled")
	}

	start := time.Now()
	result, err := AnalyzeWithRetry(ctx, logr.Discard(), primaryFn, fallbackFn)
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Less(t, elapsed, 1*time.Second, "Should return promptly when context is canceled")
}

func TestAnalyzeWithRetry_NetworkTimeout(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		if primaryCallCount == 1 {
			return nil, &mockNetError{timeout: true}
		}
		return &AnalysisResult{Content: "success after timeout"}, nil
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return &AnalysisResult{Content: "fallback"}, nil
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.NoError(t, err)
	assert.Equal(t, "success after timeout", result.Content)
	assert.Equal(t, 2, primaryCallCount)
	assert.Equal(t, 0, fallbackCallCount)
}

func TestAnalyzeWithRetry_ContextCanceledAsAPIError(t *testing.T) {
	retryDelayOverride = time.Millisecond
	t.Cleanup(func() { retryDelayOverride = 0 })

	primaryCallCount := 0
	fallbackCallCount := 0

	primaryFn := func() (*AnalysisResult, error) {
		primaryCallCount++
		return nil, context.Canceled
	}

	fallbackFn := func() (*AnalysisResult, error) {
		fallbackCallCount++
		return &AnalysisResult{Content: "fallback"}, nil
	}

	result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, primaryCallCount)
	assert.Equal(t, 0, fallbackCallCount)
}

func TestAnalyzeWithRetry_NonRetryableErrorMessages(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		sentinel error
	}{
		{"tool call failed", fmt.Errorf("%w: unknown tool", ErrToolCallFailed), ErrToolCallFailed},
		{"max iterations", ErrMaxIterations, ErrMaxIterations},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryDelayOverride = time.Millisecond
			t.Cleanup(func() { retryDelayOverride = 0 })

			primaryCallCount := 0
			fallbackCallCount := 0

			primaryFn := func() (*AnalysisResult, error) {
				primaryCallCount++
				return nil, tc.err
			}

			fallbackFn := func() (*AnalysisResult, error) {
				fallbackCallCount++
				return &AnalysisResult{Content: "fallback"}, nil
			}

			result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

			require.Error(t, err)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, tc.sentinel)
			assert.Equal(t, 1, primaryCallCount, "Should not retry for error: %v", tc.err)
			assert.Equal(t, 0, fallbackCallCount)
		})
	}
}

func TestAnalyzeWithRetry_RetryableEmptyResponseErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{"no response candidates", ErrNoResponseCandidates},
		{"no content in response", ErrNoContentInResponse},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryDelayOverride = time.Millisecond
			t.Cleanup(func() { retryDelayOverride = 0 })

			primaryCallCount := 0
			fallbackCallCount := 0

			primaryFn := func() (*AnalysisResult, error) {
				primaryCallCount++
				if primaryCallCount == 1 {
					return nil, tc.err
				}
				return &AnalysisResult{Content: "success after retry"}, nil
			}

			fallbackFn := func() (*AnalysisResult, error) {
				fallbackCallCount++
				return &AnalysisResult{Content: "fallback"}, nil
			}

			result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

			require.NoError(t, err)
			assert.Equal(t, "success after retry", result.Content)
			assert.Equal(t, 2, primaryCallCount, "Should retry for error: %v", tc.err)
			assert.Equal(t, 0, fallbackCallCount)
		})
	}
}

func TestAnalyzeWithRetry_RetryableWrappedErrors(t *testing.T) {
	testCases := []struct {
		name string
		err  error
	}{
		{
			"double wrapped retryable api error",
			fmt.Errorf("outer: %w", fmt.Errorf("gemini API error: %w", genai.APIError{Code: 429})),
		},
		{
			"triple wrapped retryable api error",
			fmt.Errorf("outer: %w", fmt.Errorf("middle: %w", fmt.Errorf("gemini API error: %w", genai.APIError{Code: 503}))),
		},
		{
			"double wrapped sentinel error",
			fmt.Errorf("outer: %w", ErrNoResponseCandidates),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			retryDelayOverride = time.Millisecond
			t.Cleanup(func() { retryDelayOverride = 0 })

			primaryCallCount := 0

			primaryFn := func() (*AnalysisResult, error) {
				primaryCallCount++
				if primaryCallCount == 1 {
					return nil, tc.err
				}
				return &AnalysisResult{Content: "success"}, nil
			}

			fallbackFn := func() (*AnalysisResult, error) {
				return &AnalysisResult{Content: "fallback"}, nil
			}

			result, err := AnalyzeWithRetry(context.Background(), logr.Discard(), primaryFn, fallbackFn)

			require.NoError(t, err)
			assert.Equal(t, "success", result.Content)
			assert.Equal(t, 2, primaryCallCount, "wrapped retryable error should trigger retry: %v", tc.err)
		})
	}
}
