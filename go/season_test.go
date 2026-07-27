package kmf1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/multiviewer"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

func newIndexOnlyTestClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()
	fixture := testfixture.Read(t, "internal/livetiming/testdata/index2024.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/static/2024/Index.json" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	c, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	tr := transport.NewWithRate(1_000_000)
	client := &Client{
		livetiming:  livetiming.NewClientWithBaseURL(tr, c, srv.URL),
		jolpica:     jolpica.NewClientWithBaseURL(tr, c, srv.URL),
		multiviewer: multiviewer.NewClientWithBaseURL(tr, c, srv.URL),
		cache:       c,
	}
	return client, srv
}

func TestSeasonEvent_RealCapture_ByRound(t *testing.T) {
	client, _ := newIndexOnlyTestClient(t)
	season := client.Year(2024)

	event, err := season.Event(context.Background(), 10)
	if err != nil {
		t.Fatalf("Event(10): %v", err)
	}
	if event.Info().EventName != "Spanish Grand Prix" {
		t.Errorf("EventName = %q, want Spanish Grand Prix", event.Info().EventName)
	}
}

func TestSeasonEvent_RealCapture_ByNameCaseInsensitive(t *testing.T) {
	client, _ := newIndexOnlyTestClient(t)
	season := client.Year(2024)

	event, err := season.Event(context.Background(), "spanish grand prix")
	if err != nil {
		t.Fatalf("Event(name): %v", err)
	}
	if event.Info().RoundNumber != 10 {
		t.Errorf("RoundNumber = %d, want 10", event.Info().RoundNumber)
	}
}

func TestSeasonEvent_NotFound(t *testing.T) {
	client, _ := newIndexOnlyTestClient(t)
	season := client.Year(2024)

	if _, err := season.Event(context.Background(), 999); err == nil {
		t.Fatal("expected an error for a non-existent round")
	}
	if _, err := season.Event(context.Background(), "not a real grand prix"); err == nil {
		t.Fatal("expected an error for a non-existent event name")
	}
}

func TestSeasonEvent_InvalidIdentifierType(t *testing.T) {
	client, _ := newIndexOnlyTestClient(t)
	season := client.Year(2024)

	if _, err := season.Event(context.Background(), 3.14); err == nil {
		t.Fatal("expected an error for an identifier that's neither int nor string")
	}
}

func TestEvent_Session_ResolvesRealPath(t *testing.T) {
	client, _ := newIndexOnlyTestClient(t)
	season := client.Year(2024)

	event, err := season.Event(context.Background(), 10)
	if err != nil {
		t.Fatalf("Event: %v", err)
	}

	ref, err := event.Session(SessionRace)
	if err != nil {
		t.Fatalf("Session(Race): %v", err)
	}
	if ref.path != "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/" {
		t.Errorf("SessionRef.path = %q", ref.path)
	}
	if ref.year != 2024 || ref.round != 10 {
		t.Errorf("SessionRef year/round = %d/%d, want 2024/10", ref.year, ref.round)
	}

	if _, err := event.Session("Not A Real Session"); err == nil {
		t.Fatal("expected an error for a non-existent session type")
	}
}

func TestEvent_Circuit_ResolvesRealCircuitKey(t *testing.T) {
	client, _ := newIndexOnlyTestClient(t)
	season := client.Year(2024)

	event, err := season.Event(context.Background(), 10)
	if err != nil {
		t.Fatalf("Event: %v", err)
	}

	circuit, err := event.Circuit()
	if err != nil {
		t.Fatalf("Circuit: %v", err)
	}
	if circuit.circuitKey != 15 {
		t.Errorf("circuitKey = %d, want 15 (Catalunya)", circuit.circuitKey)
	}
}
