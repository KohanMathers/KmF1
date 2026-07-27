package live

import (
	"sync"
	"time"
)

type TimeAnchor struct {
	mu    sync.Mutex
	start time.Time
	known bool
}

func (a *TimeAnchor) Elapsed(utcTimestamp string) time.Duration {
	if utcTimestamp == "" {
		return 0
	}
	t, err := time.Parse(time.RFC3339Nano, utcTimestamp)
	if err != nil {
		return 0
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if !a.known {
		a.start = t
		a.known = true
		return 0
	}
	return t.Sub(a.start)
}

func (a *TimeAnchor) Known() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.known
}
