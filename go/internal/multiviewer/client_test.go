package multiviewer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

func newTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()

	c, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	return &Client{http: transport.New(), cache: c, baseURL: baseURL}
}

func TestFetchCircuitInfo_RealFixtureOverHTTP(t *testing.T) {
	fixture, err := os.ReadFile("testdata/catalunya_2024.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/circuits/15/2024" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	info, err := c.FetchCircuitInfo(context.Background(), 15, 2024)
	if err != nil {
		t.Fatalf("FetchCircuitInfo: %v", err)
	}
	if len(info.Corners) != 14 {
		t.Errorf("got %d corners, want 14", len(info.Corners))
	}
}

func TestFetchCircuitInfo_404IsSurfacedAsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Circuit not found"))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)

	if _, err := c.FetchCircuitInfo(context.Background(), 999, 2024); err == nil {
		t.Fatal("expected an error for a 404 response")
	}
}

func TestFetchCircuitInfo_CachesSecondCall(t *testing.T) {
	fixture, err := os.ReadFile("testdata/catalunya_2024.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	ctx := context.Background()

	if _, err := c.FetchCircuitInfo(ctx, 15, 2024); err != nil {
		t.Fatalf("first FetchCircuitInfo: %v", err)
	}
	if _, err := c.FetchCircuitInfo(ctx, 15, 2024); err != nil {
		t.Fatalf("second FetchCircuitInfo: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("server received %d requests, want 1 (second call should hit cache)", requestCount)
	}
}
