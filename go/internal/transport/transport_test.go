package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetSetsHeadersAndReturnsBody(t *testing.T) {
	var gotUA, gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAccept = r.Header.Get("Accept-Encoding")
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Get(context.Background(), srv.URL, map[string]string{
		"User-Agent":      "KmF1-test",
		"Accept-Encoding": "identity",
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if gotUA != "KmF1-test" {
		t.Errorf("server saw User-Agent %q, want %q", gotUA, "KmF1-test")
	}
	if gotAccept != "identity" {
		t.Errorf("server saw Accept-Encoding %q, want %q", gotAccept, "identity")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(body) != "hello" {
		t.Fatalf("got body %q, want %q", body, "hello")
	}
}

func TestGetPropagatesNon2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Get(context.Background(), srv.URL, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestGetRespectsContextCancellation(t *testing.T) {
	c := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := c.Get(ctx, "http://127.0.0.1:0/unused", nil); err == nil {
		t.Fatal("expected an error for an already-cancelled context")
	}
}

func TestGetRejectsMalformedURL(t *testing.T) {
	c := New()
	if _, err := c.Get(context.Background(), "://not-a-url", nil); err == nil {
		t.Fatal("expected an error for a malformed URL")
	}
}

func TestRateLimitAllowsBurst(t *testing.T) {
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	for i := 0; i < burst; i++ {
		if err := c.limiterFor("example.com").Wait(ctx); err != nil {
			t.Fatalf("request %d within burst should not be rate limited: %v", i, err)
		}
	}
}

func TestRateLimitBlocksBeyondBurst(t *testing.T) {
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	for i := 0; i < burst; i++ {
		if err := c.limiterFor("example.com").Wait(ctx); err != nil {
			t.Fatalf("request %d within burst should not be rate limited: %v", i, err)
		}
	}

	if err := c.limiterFor("example.com").Wait(ctx); err == nil {
		t.Fatal("expected request beyond burst to be rate limited")
	}
}

func TestRateLimitIsPerHost(t *testing.T) {
	c := New()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	for i := 0; i < burst; i++ {
		if err := c.limiterFor("a.example.com").Wait(ctx); err != nil {
			t.Fatalf("host a request %d should not be rate limited: %v", i, err)
		}
	}

	if err := c.limiterFor("b.example.com").Wait(ctx); err != nil {
		t.Fatalf("host b should have its own independent limit: %v", err)
	}
}

func TestNew_SetsDefaultTimeout(t *testing.T) {
	c := New()
	if c.http.Timeout != defaultTimeout {
		t.Errorf("http.Client.Timeout = %v, want %v", c.http.Timeout, defaultTimeout)
	}
}

func withFastRetries(t *testing.T) {
	t.Helper()
	orig := retryBaseDelay
	retryBaseDelay = time.Millisecond
	t.Cleanup(func() { retryBaseDelay = orig })
}

func TestGet_RetriesTransientFailureThenSucceeds(t *testing.T) {
	withFastRetries(t)

	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) <= 2 {

			hj, ok := w.(http.Hijacker)
			if !ok {
				t.Fatal("ResponseWriter does not support hijacking")
			}
			conn, _, err := hj.Hijack()
			if err != nil {
				t.Fatalf("hijack: %v", err)
			}
			conn.Close()
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Get(context.Background(), srv.URL, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 3 {
		t.Fatalf("server saw %d attempts, want 3 (2 transient failures + 1 success)", got)
	}
}

func TestGet_GivesUpAfterMaxRetries(t *testing.T) {
	withFastRetries(t)

	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		hj, _ := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	c := New()
	if _, err := c.Get(context.Background(), srv.URL, nil); err == nil {
		t.Fatal("expected an error once every attempt fails transiently")
	}

	if got := atomic.LoadInt32(&attempts); got != maxRetries+1 {
		t.Fatalf("server saw %d attempts, want %d (initial + %d retries)", got, maxRetries+1, maxRetries)
	}
}

func TestGet_DoesNotRetryNon2xxStatus(t *testing.T) {
	withFastRetries(t)

	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New()
	resp, err := c.Get(context.Background(), srv.URL, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer resp.Body.Close()

	if got := atomic.LoadInt32(&attempts); got != 1 {
		t.Fatalf("server saw %d attempts, want 1 (a 5xx status is not a transport-level retry case)", got)
	}
}

func TestGet_StopsRetryingOnContextCancellation(t *testing.T) {
	withFastRetries(t)
	retryBaseDelay = 200 * time.Millisecond

	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		hj, _ := w.(http.Hijacker)
		conn, _, _ := hj.Hijack()
		conn.Close()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := New()
	start := time.Now()
	if _, err := c.Get(ctx, srv.URL, nil); err == nil {
		t.Fatal("expected an error")
	}
	if elapsed := time.Since(start); elapsed > 150*time.Millisecond {
		t.Errorf("Get took %v, want it to stop quickly once the context expired mid-backoff", elapsed)
	}
}

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, false},
		{"wrapped context canceled", fmt.Errorf("request: %w", context.Canceled), false},
		{"plain non-network error", errors.New("boom"), false},
		{"timeout net.Error", &timeoutError{}, true},
		{"net.OpError (connection refused etc.)", &net.OpError{Op: "dial", Err: errors.New("refused")}, true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"plain EOF", io.EOF, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isRetryable(tc.err); got != tc.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }
