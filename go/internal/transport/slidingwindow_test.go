package transport

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSlidingWindowLimiter_GrantsUpToLimitInstantly(t *testing.T) {
	l := newSlidingWindowLimiter(5, time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	for i := 0; i < 5; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("grant %d: %v", i, err)
		}
	}
	if elapsed := time.Since(start); elapsed > 10*time.Millisecond {
		t.Errorf("5 grants against a limit of 5 took %v, want near-instant", elapsed)
	}
}

func TestSlidingWindowLimiter_BlocksOnceAtLimit(t *testing.T) {
	l := newSlidingWindowLimiter(2, time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	for i := 0; i < 2; i++ {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("grant %d: %v", i, err)
		}
	}
	if err := l.Wait(ctx); err == nil {
		t.Fatal("expected the 3rd request to block until the context expired")
	}
}

func TestSlidingWindowLimiter_GrantsAgainOnceOldestAgesOut(t *testing.T) {
	window := 60 * time.Millisecond
	l := newSlidingWindowLimiter(1, window)
	ctx := context.Background()

	start := time.Now()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("first grant: %v", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, window*3)
	defer cancel()
	if err := l.Wait(waitCtx); err != nil {
		t.Fatalf("second grant should succeed once the window elapses: %v", err)
	}

	if elapsed := time.Since(start); elapsed < window {
		t.Errorf("second grant arrived after %v, want at least the %v window", elapsed, window)
	}
}

func TestSlidingWindowLimiter_IsSafeForConcurrentUse(t *testing.T) {
	l := newSlidingWindowLimiter(50, time.Hour)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var granted int
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := l.Wait(ctx); err == nil {
				mu.Lock()
				granted++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if granted != 50 {
		t.Errorf("granted = %d, want 50 (all within the limit)", granted)
	}
}

func TestSlidingWindowLimiter_RespectsContextCancellation(t *testing.T) {
	l := newSlidingWindowLimiter(1, time.Hour)
	ctx := context.Background()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("first grant: %v", err)
	}

	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if err := l.Wait(cancelCtx); err == nil {
		t.Fatal("expected an error for an already-cancelled context")
	}
}
