package cache

import (
	"testing"
	"time"
)

func open(t *testing.T) *Cache {
	t.Helper()
	c, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestGetMiss(t *testing.T) {
	c := open(t)

	if _, found, err := c.Get("https://example.com/missing", 1); err != nil {
		t.Fatalf("Get: %v", err)
	} else if found {
		t.Fatal("expected a miss for an entry that was never set")
	}
}

func TestSetThenGet(t *testing.T) {
	c := open(t)
	const url = "https://livetiming.formula1.com/static/example/SessionInfo.jsonStream"

	if err := c.Set(url, 1, []byte("hello"), time.Time{}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	body, found, err := c.Get(url, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("expected a hit after Set")
	}
	if string(body) != "hello" {
		t.Fatalf("got body %q, want %q", body, "hello")
	}
}

func TestDifferentParserVersionsAreIndependent(t *testing.T) {
	c := open(t)
	const url = "https://api.jolpi.ca/ergast/f1/2024/11/sprint.json"

	if err := c.Set(url, 1, []byte("v1"), time.Time{}); err != nil {
		t.Fatalf("Set v1: %v", err)
	}
	if err := c.Set(url, 2, []byte("v2"), time.Time{}); err != nil {
		t.Fatalf("Set v2: %v", err)
	}

	body, found, err := c.Get(url, 1)
	if err != nil || !found {
		t.Fatalf("Get v1: body=%q found=%v err=%v", body, found, err)
	}
	if string(body) != "v1" {
		t.Fatalf("v1 got %q, want %q", body, "v1")
	}

	body, found, err = c.Get(url, 2)
	if err != nil || !found {
		t.Fatalf("Get v2: body=%q found=%v err=%v", body, found, err)
	}
	if string(body) != "v2" {
		t.Fatalf("v2 got %q, want %q", body, "v2")
	}
}

func TestPermanentEntryNeverExpires(t *testing.T) {
	c := open(t)
	const url = "https://example.com/finalized-archive"

	if err := c.Set(url, 1, []byte("final"), time.Time{}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, found, err := c.Get(url, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !found {
		t.Fatal("a permanent entry (zero expiresAt) should never be reported as expired")
	}
}

func TestExpiredEntryIsATreatedAsAMiss(t *testing.T) {
	c := open(t)
	const url = "https://example.com/in-progress-session"

	if err := c.Set(url, 1, []byte("stale"), time.Now().Add(-time.Minute)); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_, found, err := c.Get(url, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if found {
		t.Fatal("expected an already-expired entry to be reported as a miss")
	}
}

func TestSetOverwritesExistingEntry(t *testing.T) {
	c := open(t)
	const url = "https://example.com/overwrite-me"

	if err := c.Set(url, 1, []byte("old"), time.Time{}); err != nil {
		t.Fatalf("Set old: %v", err)
	}
	if err := c.Set(url, 1, []byte("new"), time.Time{}); err != nil {
		t.Fatalf("Set new: %v", err)
	}

	body, found, err := c.Get(url, 1)
	if err != nil || !found {
		t.Fatalf("Get: body=%q found=%v err=%v", body, found, err)
	}
	if string(body) != "new" {
		t.Fatalf("got %q, want %q after overwrite", body, "new")
	}
}
