package kmf1

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/multiviewer"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

func TestSessionRefLoad_FeedFailure_LogsWarning(t *testing.T) {
	fixturesMinusTrackStatus := make(map[string]string, len(livetimingFixtures))
	for feed, fixture := range livetimingFixtures {
		if feed == "TrackStatus.jsonStream" {
			continue
		}
		fixturesMinusTrackStatus[feed] = fixture
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			fixture, ok := fixturesMinusTrackStatus[path.Base(r.URL.Path)]
			if !ok {
				http.NotFound(w, r)
				return
			}
			data, err := os.ReadFile(fixture)
			if err != nil {
				t.Errorf("read fixture %s: %v", fixture, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(data)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	defer func() { _ = c.Close() }()

	tr := transport.NewWithRate(1_000_000, 1000)
	var logs bytes.Buffer
	client := &Client{
		livetiming:  livetiming.NewClientWithBaseURL(tr, c, srv.URL),
		jolpica:     jolpica.NewClientWithBaseURL(tr, c, srv.URL),
		multiviewer: multiviewer.NewClientWithBaseURL(tr, c, srv.URL),
		cache:       c,
		logger:      slog.New(slog.NewTextHandler(&logs, nil)),
	}
	ref := &SessionRef{
		client:      client,
		year:        2024,
		round:       9,
		sessionType: SessionRace,
		path:        "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/",
	}

	session, err := ref.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v (should still succeed — TrackStatus is best-effort)", err)
	}
	if len(session.TrackStatus()) != 0 {
		t.Errorf("TrackStatus() = %v, want empty (fixture withheld)", session.TrackStatus())
	}
	if len(session.Drivers()) == 0 {
		t.Error("expected a non-empty driver roster (DriverList fixture was provided)")
	}

	out := logs.String()
	if !strings.Contains(out, "feed fetch failed") {
		t.Errorf("expected a fetch-failure log line, got:\n%s", out)
	}
	if !strings.Contains(out, "level=WARN") || !strings.Contains(out, "feed=TrackStatus") {
		t.Errorf("expected a Warn-level log line naming the TrackStatus feed, got:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "feed=DriverList") && strings.Contains(line, "level=WARN") {
			t.Errorf("expected no failure logged for DriverList (its fixture was provided), got line:\n%s", line)
		}
	}
	if !strings.Contains(out, `msg="kmf1: feed loaded" feed=DriverList`) {
		t.Errorf("expected an Info success line for DriverList (its fixture was provided), got:\n%s", out)
	}
	if !strings.Contains(out, `msg="kmf1: loading session"`) {
		t.Errorf("expected an Info line marking the start of Load, got:\n%s", out)
	}
	if !strings.Contains(out, `msg="kmf1: session loaded"`) {
		t.Errorf("expected an Info line marking the end of Load, got:\n%s", out)
	}
}

func TestClient_Log_DefaultsToDiscard(t *testing.T) {
	var c Client
	logger := c.log()
	if logger == nil {
		t.Fatal("log() returned nil, want a non-nil discarding logger")
	}
	logger.Warn("should not panic and should go nowhere")
}

func TestWithLogger_ConfiguresClient(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	client, err := New(t.TempDir()+"/cache.sqlite", WithLogger(logger))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = client.Close() }()

	if client.log() != logger {
		t.Error("client.log() did not return the logger passed via WithLogger")
	}
}
