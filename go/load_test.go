package kmf1

import (
	"context"
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

var livetimingFixtures = map[string]string{
	"SessionInfo.jsonStream":            "internal/livetiming/feeds/testdata/sessioninfo.jsonstream",
	"DriverList.jsonStream":             "internal/livetiming/feeds/testdata/driverlist.jsonstream",
	"TrackStatus.jsonStream":            "internal/livetiming/feeds/testdata/trackstatus.jsonstream",
	"SessionStatus.jsonStream":          "internal/livetiming/feeds/testdata/sessionstatus.jsonstream",
	"LapCount.jsonStream":               "internal/livetiming/feeds/testdata/lapcount.jsonstream",
	"TimingData.jsonStream":             "internal/livetiming/feeds/testdata/timingdata.jsonstream",
	"TimingAppData.jsonStream":          "internal/livetiming/feeds/testdata/timingappdata.jsonstream",
	"CarData.z.jsonStream":              "internal/livetiming/feeds/testdata/cardata.jsonstream",
	"Position.z.jsonStream":             "internal/livetiming/feeds/testdata/position.jsonstream",
	"WeatherData.jsonStream":            "internal/livetiming/feeds/testdata/weatherdata.jsonstream",
	"RaceControlMessages.jsonStream":    "internal/livetiming/feeds/testdata/racecontrolmessages.jsonstream",
	"TimingStats.jsonStream":            "internal/livetiming/feeds/testdata/timingstats.jsonstream",
	"TeamRadio.jsonStream":              "internal/livetiming/feeds/testdata/teamradio.jsonstream",
	"TopThree.jsonStream":               "internal/livetiming/feeds/testdata/topthree.jsonstream",
	"ChampionshipPrediction.jsonStream": "internal/livetiming/feeds/testdata/championshipprediction.jsonstream",
	"ExtrapolatedClock.jsonStream":      "internal/livetiming/feeds/testdata/extrapolatedclock.jsonstream",
	"ArchiveStatus.json":                "internal/livetiming/feeds/testdata/archivestatus.json",
	"AudioStreams.jsonStream":           "internal/livetiming/feeds/testdata/audiostreams.jsonstream",
	"ContentStreams.jsonStream":         "internal/livetiming/feeds/testdata/contentstreams.jsonstream",
}

var jolpicaFixtures = map[string]string{
	"/2024/9/results.json":  "internal/jolpica/testdata/results_2024_9.json",
	"/2024/9/pitstops.json": "internal/jolpica/testdata/pitstops_2024_9.json",
	"/2024/9/laps/1.json":   "internal/jolpica/testdata/laps1_2024_9.json",
}

func newFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/") {
			feedFile := path.Base(r.URL.Path)
			fixture, ok := livetimingFixtures[feedFile]
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

		if fixture, ok := jolpicaFixtures[r.URL.Path]; ok {
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
}

func newFixtureTestClient(t *testing.T, baseURL string) *Client {
	t.Helper()
	c, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })

	tr := transport.NewWithRate(1_000_000, 1000)
	return &Client{
		livetiming:  livetiming.NewClientWithBaseURL(tr, c, baseURL),
		jolpica:     jolpica.NewClientWithBaseURL(tr, c, baseURL),
		multiviewer: multiviewer.NewClientWithBaseURL(tr, c, baseURL),
		cache:       c,
	}
}

func TestSessionRefLoad_RealFixtures_EndToEnd(t *testing.T) {
	srv := newFixtureServer(t)
	defer srv.Close()

	client := newFixtureTestClient(t, srv.URL)
	ref := &SessionRef{
		client:      client,
		year:        2024,
		round:       9,
		sessionType: SessionRace,
		path:        "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/",
	}

	session, err := ref.Load(context.Background())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if session.Info() == nil || session.Info().Meeting.Name != "Spanish Grand Prix" {
		t.Errorf("Info() = %+v, want the real Spanish GP SessionInfo", session.Info())
	}

	drivers := session.Drivers()
	if len(drivers) == 0 {
		t.Fatal("expected a non-empty driver roster")
	}

	laps := session.Laps()
	if len(laps) == 0 {
		t.Fatal("expected non-empty Laps() from the real TimingData fixture")
	}

	if len(session.Weather()) == 0 {
		t.Error("expected non-empty Weather()")
	}
	if len(session.TrackStatus()) == 0 {
		t.Error("expected non-empty TrackStatus()")
	}
	if len(session.RaceControlMessages()) == 0 {
		t.Error("expected non-empty RaceControlMessages()")
	}
	if len(session.TeamRadio()) == 0 {
		t.Error("expected non-empty TeamRadio()")
	}
	if len(session.TopThree()) == 0 {
		t.Error("expected non-empty TopThree()")
	}
	if len(session.BestStats()) == 0 {
		t.Error("expected non-empty BestStats()")
	}
	if session.ArchiveStatus().Status == "" {
		t.Error("expected a non-empty ArchiveStatus")
	}
	if len(session.Clock()) == 0 {
		t.Error("expected non-empty Clock()")
	}
	if len(session.Broadcast().Audio()) == 0 {
		t.Error("expected non-empty Broadcast().Audio()")
	}
	if len(session.Broadcast().Content()) == 0 {
		t.Error("expected non-empty Broadcast().Content()")
	}

	cp := session.ChampionshipPrediction()
	if cp == nil {
		t.Fatal("expected a non-nil ChampionshipPrediction for a Race session")
	}
	if len(cp.Drivers()) == 0 {
		t.Error("expected non-empty ChampionshipPrediction().Drivers()")
	}

	if len(session.Results()) == 0 {
		t.Fatal("expected non-empty Results() from Jolpica reconciliation")
	}
	var anyPosition bool
	for _, r := range session.Results() {
		if r.Position != nil {
			anyPosition = true
		}
	}
	if !anyPosition {
		t.Error("expected at least one result with a real Position from Jolpica")
	}

	if len(session.PitStops()) == 0 {
		t.Error("expected non-empty PitStops() (Race session, real Jolpica pit stop data)")
	}

	ver := session.Driver("1")
	if !ver.Exists() {
		t.Fatal("expected driver 1 (VER) to exist")
	}
	if ver.Info().Abbreviation != "VER" {
		t.Errorf("driver 1 Abbreviation = %q, want VER", ver.Info().Abbreviation)
	}
	verByAbbr := session.Driver("ver")
	if !verByAbbr.Exists() || verByAbbr.Info().DriverNumber != "1" {
		t.Errorf("Driver(\"ver\") = %+v, want driver 1", verByAbbr.Info())
	}

	if len(ver.Laps()) == 0 {
		t.Error("expected driver 1 to have laps")
	}
	if len(ver.Telemetry()) == 0 {
		t.Error("expected driver 1 to have telemetry")
	}

	notFound := session.Driver("does-not-exist")
	if notFound.Exists() {
		t.Error("expected Driver(\"does-not-exist\") to not exist")
	}
	if len(notFound.Laps()) != 0 {
		t.Error("expected a not-found Driver's Laps() to be empty, not panic")
	}
	if _, ok := notFound.Result(); ok {
		t.Error("expected a not-found Driver's Result() to report ok=false")
	}
}
