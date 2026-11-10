package jolpica

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

func newGlobalRefTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return &Client{http: transport.New(), cache: c, baseURL: baseURL}
}

func TestFetchDriverReference_RealCapture_NotSeasonScoped(t *testing.T) {
	fixture := testfixture.Read(t, "testdata/driver_max_verstappen.json")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/drivers/max_verstappen.json" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	driver, err := c.FetchDriverReference(context.Background(), "max_verstappen")
	if err != nil {
		t.Fatalf("FetchDriverReference: %v", err)
	}
	if driver.Abbreviation != "VER" {
		t.Errorf("Abbreviation = %q, want VER", driver.Abbreviation)
	}
}

func TestFetchConstructorReference_RealCapture(t *testing.T) {
	fixture := testfixture.Read(t, "testdata/constructor_red_bull.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/constructors/red_bull.json" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	constructor, err := c.FetchConstructorReference(context.Background(), "red_bull")
	if err != nil {
		t.Fatalf("FetchConstructorReference: %v", err)
	}
	if constructor.Name != "Red Bull" {
		t.Errorf("Name = %q, want Red Bull", constructor.Name)
	}
}

func TestFetchCircuitReference_RealCapture(t *testing.T) {
	fixture := testfixture.Read(t, "testdata/circuit_monza.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/circuits/monza.json" {
			t.Errorf("unexpected request path %q", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	circuit, err := c.FetchCircuitReference(context.Background(), "monza")
	if err != nil {
		t.Fatalf("FetchCircuitReference: %v", err)
	}
	if circuit.Name != "Autodromo Nazionale di Monza" {
		t.Errorf("Name = %q", circuit.Name)
	}
}

func TestFetchDriverReference_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"MRData":{"DriverTable":{"Drivers":[]}}}`))
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	if _, err := c.FetchDriverReference(context.Background(), "nobody"); err == nil {
		t.Fatal("expected an error for a driverId with no matching driver")
	}
}

func TestFetchAllFinishingStatusReferences_RealCapture(t *testing.T) {
	fixture := testfixture.Read(t, "testdata/status_alltime.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status.json" {
			t.Errorf("unexpected request path %q, want /status.json (not year-scoped)", r.URL.Path)
		}
		if r.URL.Query().Get("offset") != "0" {
			_, _ = w.Write(syntheticStatusPage(100, 36, 136))
			return
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	statuses, err := c.FetchAllFinishingStatusReferences(context.Background())
	if err != nil {
		t.Fatalf("FetchAllFinishingStatusReferences: %v", err)
	}
	if len(statuses) == 0 {
		t.Fatal("expected at least one status")
	}
	if len(statuses) != 136 {
		t.Errorf("got %d statuses, want 136 (all pages)", len(statuses))
	}
	if statuses[0].Status != "Finished" || statuses[0].Count != 8117 {
		t.Errorf("first status = %+v, want Finished/8117", statuses[0])
	}
}

func TestFetchAllDriverReferences_UsesUnscopedPath(t *testing.T) {
	fixture := testfixture.Read(t, "testdata/drivers_2024.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/drivers.json" {
			t.Errorf("unexpected request path %q, want /drivers.json (not year-scoped)", r.URL.Path)
		}
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	drivers, err := c.FetchAllDriverReferences(context.Background())
	if err != nil {
		t.Fatalf("FetchAllDriverReferences: %v", err)
	}
	if len(drivers) != 25 {
		t.Errorf("got %d drivers, want 25", len(drivers))
	}
}
