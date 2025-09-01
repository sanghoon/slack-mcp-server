package provider

import (
	"context"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestEnforceRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		rateLimiter *rate.Limiter
		expectError bool
	}{
		{
			name:        "nil rate limiter",
			rateLimiter: nil,
			expectError: false,
		},
		{
			name:        "rate limiter allows",
			rateLimiter: rate.NewLimiter(rate.Every(100*time.Millisecond), 1),
			expectError: false,
		},
		{
			name:        "rate limiter with cancelled context",
			rateLimiter: rate.NewLimiter(rate.Every(10*time.Second), 1),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := &ApiProvider{
				rateLimiter: tt.rateLimiter,
			}

			ctx := context.Background()
			if tt.expectError {
				// Use cancelled context to simulate rate limit wait error
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelCtx
			}

			err := ap.enforceRateLimit(ctx)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestRateLimitedAPICalls(t *testing.T) {
	// Create a rate limiter that allows 2 requests per second
	limiter := rate.NewLimiter(rate.Every(500*time.Millisecond), 2)
	
	ap := &ApiProvider{
		rateLimiter: limiter,
	}

	// Test that we can make 2 calls immediately
	ctx := context.Background()
	
	start := time.Now()
	
	// First call should succeed immediately
	err := ap.enforceRateLimit(ctx)
	if err != nil {
		t.Errorf("first call failed: %v", err)
	}
	
	// Second call should succeed immediately (burst of 2)
	err = ap.enforceRateLimit(ctx)
	if err != nil {
		t.Errorf("second call failed: %v", err)
	}
	
	// Third call should be rate limited and take ~500ms
	err = ap.enforceRateLimit(ctx)
	if err != nil {
		t.Errorf("third call failed: %v", err)
	}
	
	elapsed := time.Since(start)
	if elapsed < 400*time.Millisecond {
		t.Errorf("rate limiting not working: elapsed time %v, expected >= 400ms", elapsed)
	}
}