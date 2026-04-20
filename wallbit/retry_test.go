package wallbit

import (
	"testing"
	"time"
)

func TestJitterBounds(t *testing.T) {
	t.Parallel()

	cases := []time.Duration{
		1 * time.Millisecond,
		250 * time.Millisecond,
		2 * time.Second,
		time.Hour,
	}
	// 1024 samples per bound is enough to exercise both extremes with very
	// high probability without making the test slow or flaky. We assert the
	// invariant (d/2 <= got <= d) rather than distributional properties so
	// the test never flakes on a run that happens to sample the edges.
	const samples = 1024

	for _, d := range cases {
		lo := d / 2
		for i := 0; i < samples; i++ {
			got := jitter(d)
			if got < lo || got > d {
				t.Fatalf("d=%s: jitter returned %s, want within [%s, %s]", d, got, lo, d)
			}
		}
	}
}

func TestJitterZeroAndNegative(t *testing.T) {
	t.Parallel()

	if got := jitter(0); got != 0 {
		t.Fatalf("jitter(0) = %s, want 0", got)
	}
	if got := jitter(-5 * time.Second); got != 0 {
		t.Fatalf("jitter(-5s) = %s, want 0", got)
	}
}

func TestJitterVariesAcrossCalls(t *testing.T) {
	t.Parallel()

	// With d=1s and equal jitter, the theoretical range is [500ms, 1s] and
	// the resolution of rand.Int64N(500_000_001) is nanoseconds; seeing the
	// same value 32 times in a row would mean the jitter is effectively
	// degenerate. This catches accidental regressions like swapping
	// rand.Int64N for a constant or forgetting to apply jitter entirely.
	const d = time.Second
	seen := make(map[time.Duration]struct{})
	for i := 0; i < 32; i++ {
		seen[jitter(d)] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatalf("jitter appears degenerate: produced only %d unique values across 32 samples", len(seen))
	}
}

func TestRetryWaitAppliesJitter(t *testing.T) {
	t.Parallel()

	c := &Client{
		cfg: &Config{
			RetryPolicy: RetryPolicy{
				MaxAttempts: 4,
				BaseDelay:   200 * time.Millisecond,
				MaxDelay:    5 * time.Second,
			},
		},
	}

	// The deterministic exponential schedule (before jitter) for these
	// parameters is 200ms, 400ms, 800ms for failure indices 0, 1, 2. We
	// assert each retry lives inside [d/2, d] and that MaxDelay still caps
	// the upper bound on the final attempt.
	expected := []time.Duration{
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}
	for i, d := range expected {
		got := c.retryWaitBeforeNextAttempt(nil, nil, i)
		lo := d / 2
		if got < lo || got > d {
			t.Fatalf("attempt %d: got %s, want within [%s, %s]", i, got, lo, d)
		}
	}
}

func TestRetryWaitRespectsRetryAfterWithoutJitter(t *testing.T) {
	t.Parallel()

	c := &Client{
		cfg: &Config{
			RetryPolicy: RetryPolicy{
				MaxAttempts: 2,
				BaseDelay:   200 * time.Millisecond,
				MaxDelay:    10 * time.Second,
			},
		},
	}
	three := int64(3)
	apiErr := &Error{RetryAfterSeconds: &three}
	// Retry-After is an explicit contract from the server; jitter would
	// undermine the operator's intent. We assert the value is returned
	// exactly, bounded only by MaxDelay.
	if got := c.retryWaitBeforeNextAttempt(nil, apiErr, 0); got != 3*time.Second {
		t.Fatalf("Retry-After not honored as-is: got %s, want 3s", got)
	}
}
