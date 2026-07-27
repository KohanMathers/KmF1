package transport

import (
	"context"
	"sync"
	"time"
)

type slidingWindowLimiter struct {
	limit  int
	window time.Duration

	mu    sync.Mutex
	times []time.Time
}

func newSlidingWindowLimiter(limit int, window time.Duration) *slidingWindowLimiter {
	return &slidingWindowLimiter{limit: limit, window: window}
}

func (l *slidingWindowLimiter) Wait(ctx context.Context) error {
	for {
		wait, ok := l.tryAcquire(time.Now())
		if ok {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			// A slot may have freed up; loop around and recheck.
		}
	}
}

func (l *slidingWindowLimiter) tryAcquire(now time.Time) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	i := 0
	for i < len(l.times) && l.times[i].Before(cutoff) {
		i++
	}
	l.times = l.times[i:]

	if len(l.times) < l.limit {
		l.times = append(l.times, now)
		return 0, true
	}

	wait := l.times[0].Add(l.window).Sub(now)
	if wait < 0 {
		wait = 0
	}
	return wait, false
}
