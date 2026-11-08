package livetiming

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

func newTestClient(t *testing.T, primaryURL, mirrorURL string) *Client {
	t.Helper()

	c, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	return &Client{
		http:          transport.New(),
		cache:         c,
		baseURL:       primaryURL,
		mirrorBaseURL: mirrorURL,
	}
}

func TestFetchSeason_RealIndexFixture(t *testing.T) {
	fixture := testfixture.Read(t, "testdata/index2024.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/2024/Index.json" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, "http://unused.invalid")

	season, err := c.FetchSeason(context.Background(), 2024)
	if err != nil {
		t.Fatalf("FetchSeason: %v", err)
	}

	if season.Year != 2024 {
		t.Errorf("Year = %d, want 2024", season.Year)
	}
	if len(season.Meetings) != 15 {
		t.Errorf("got %d meetings, want 15", len(season.Meetings))
	}

	var spanish *Meeting
	for i := range season.Meetings {
		if season.Meetings[i].Name == "Spanish Grand Prix" {
			spanish = &season.Meetings[i]
			break
		}
	}
	if spanish == nil {
		t.Fatal("Spanish Grand Prix not found in parsed real fixture")
	}
	if len(spanish.Sessions) != 5 {
		t.Fatalf("got %d Spanish GP sessions, want 5", len(spanish.Sessions))
	}

	var race *Session
	for i := range spanish.Sessions {
		if spanish.Sessions[i].Name == "Race" {
			race = &spanish.Sessions[i]
		}
	}
	if race == nil {
		t.Fatal("Spanish GP Race session not found")
	}

	const wantPath = "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/"
	if race.Path != wantPath {
		t.Errorf("Race.Path = %q, want %q", race.Path, wantPath)
	}
}

func TestFetchFeed_ConstructsCorrectURL(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, "http://unused.invalid")

	body, err := c.FetchFeed(context.Background(),
		"2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/", "SessionInfo.jsonStream")
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}

	const wantPath = "/static/2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/SessionInfo.jsonStream"
	if gotPath != wantPath {
		t.Errorf("request path = %q, want %q", gotPath, wantPath)
	}
	if string(body) != `{"ok":true}` {
		t.Errorf("got body %q", body)
	}
}

func TestFetch_FallsBackToMirrorOnPrimaryFailure(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"from":"mirror"}`))
	}))
	defer mirror.Close()

	c := newTestClient(t, primary.URL, mirror.URL)

	body, err := c.FetchFeed(context.Background(), "2024/example/", "SessionInfo.jsonStream")
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if string(body) != `{"from":"mirror"}` {
		t.Errorf("got body %q, want the mirror's response", body)
	}
}

func TestFetch_ErrorsWhenBothHostsFail(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer mirror.Close()

	c := newTestClient(t, primary.URL, mirror.URL)

	if _, err := c.FetchFeed(context.Background(), "2024/example/", "SessionInfo.jsonStream"); err == nil {
		t.Fatal("expected an error when both primary and mirror fail")
	}
}

func TestFetch_CachesSuccessfulResponses(t *testing.T) {
	var requests int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&requests, 1)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, "http://unused.invalid")
	ctx := context.Background()

	if _, err := c.FetchFeed(ctx, "2024/example/", "SessionInfo.jsonStream"); err != nil {
		t.Fatalf("first FetchFeed: %v", err)
	}
	if _, err := c.FetchFeed(ctx, "2024/example/", "SessionInfo.jsonStream"); err != nil {
		t.Fatalf("second FetchFeed: %v", err)
	}

	if got := atomic.LoadInt32(&requests); got != 1 {
		t.Errorf("server received %d requests, want 1 (second call should be a cache hit)", got)
	}
}

func TestGet_SendsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, "http://unused.invalid")
	if _, err := c.FetchFeed(context.Background(), "2024/example/", "SessionInfo.jsonStream"); err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}

	if gotUA != userAgent {
		t.Errorf("User-Agent = %q, want %q", gotUA, userAgent)
	}
}

func TestFetch_ReportsAccessDeniedFromPrimary(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer primary.Close()

	mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mirror.Close()

	c := newTestClient(t, primary.URL, mirror.URL)

	_, err := c.FetchSeason(context.Background(), 2023)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !IsAccessDenied(err) {
		t.Errorf("IsAccessDenied = false for %v", err)
	}
	for _, want := range []string{"status 401", "status 404"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestFetch_ServesExpiredCacheWhenUpstreamFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL, srv.URL)
	const path = "/static/2022/x/2022-03-27_Race/SessionInfo.jsonStream"
	if err := c.cache.Set(srv.URL+path, parserVersion, []byte("cached"), time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("cache.Set: %v", err)
	}

	body, err := c.FetchFeed(context.Background(), "2022/x/2022-03-27_Race/", "SessionInfo.jsonStream")
	if err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if string(body) != "cached" {
		t.Errorf("got body %q, want the stale cached body", body)
	}
}

func TestExpiryFor(t *testing.T) {
	now := time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		path      string
		permanent bool
	}{
		{"/static/2022/Index.json", true},
		{"/static/2026/Index.json", false},
		{"/static/2024/2024-03-02_Bahrain_Grand_Prix/2024-03-02_Race/SessionInfo.jsonStream", true},
		{"/static/2026/2026-11-08_Brazilian_Grand_Prix/2026-11-08_Race/SessionInfo.jsonStream", false},
		{"/static/StreamingStatus.json", false},
	}
	for _, tc := range cases {
		got := expiryFor(tc.path, now).IsZero()
		if got != tc.permanent {
			t.Errorf("expiryFor(%q) permanent = %v, want %v", tc.path, got, tc.permanent)
		}
	}
}

func TestFetch_SendsAuthTokenToPrimaryOnly(t *testing.T) {
	var primaryAuth, mirrorAuth string
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		primaryAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer primary.Close()

	mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mirrorAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte("ok"))
	}))
	defer mirror.Close()

	c := newTestClient(t, primary.URL, mirror.URL)
	c.SetAuthToken("secret")

	if _, err := c.FetchFeed(context.Background(), "2024/example/", "SessionInfo.jsonStream"); err != nil {
		t.Fatalf("FetchFeed: %v", err)
	}
	if primaryAuth != "Bearer secret" {
		t.Errorf("primary Authorization = %q, want %q", primaryAuth, "Bearer secret")
	}
	if mirrorAuth != "" {
		t.Errorf("mirror received Authorization %q, want none", mirrorAuth)
	}
}
