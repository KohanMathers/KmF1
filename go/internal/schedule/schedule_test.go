package schedule

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func loadRealSeason(t *testing.T) *livetiming.Season {
	t.Helper()
	data := testfixture.Read(t, "../livetiming/testdata/index2024.json")

	var season livetiming.Season
	if err := json.Unmarshal(wire.TrimBOM(data), &season); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	season.Year = 2024
	return &season
}

func TestScheduleFromLivetiming_RealCapture_ConventionalWeekend(t *testing.T) {
	season := loadRealSeason(t)
	schedules := scheduleFromLivetiming(season)

	var found bool
	for _, s := range schedules {
		if s.EventName != "Spanish Grand Prix" {
			continue
		}
		found = true

		if s.Season != 2024 || s.RoundNumber != 10 {
			t.Errorf("Season/RoundNumber = %d/%d, want 2024/10", s.Season, s.RoundNumber)
		}
		if s.OfficialEventName != "FORMULA 1 ARAMCO GRAN PREMIO DE ESPAÑA 2024" {
			t.Errorf("OfficialEventName = %q", s.OfficialEventName)
		}
		if s.Country != "Spain" || s.Location != "Barcelona" {
			t.Errorf("Country/Location = %q/%q", s.Country, s.Location)
		}
		if s.EventFormat != "conventional" {
			t.Errorf("EventFormat = %q, want conventional", s.EventFormat)
		}
		if s.CircuitKey != 15 {
			t.Errorf("CircuitKey = %d, want 15 (Catalunya)", s.CircuitKey)
		}

		wantTypes := []string{"Practice 1", "Practice 2", "Practice 3", "Qualifying", "Race"}
		if len(s.Sessions) != len(wantTypes) {
			t.Fatalf("got %d sessions, want %d: %+v", len(s.Sessions), len(wantTypes), s.Sessions)
		}
		for i, want := range wantTypes {
			if s.Sessions[i].SessionType != want {
				t.Errorf("session %d = %q, want %q", i, s.Sessions[i].SessionType, want)
			}
		}

		race := s.Sessions[len(s.Sessions)-1]
		wantLocal := time.Date(2024, 6, 23, 15, 0, 0, 0, time.FixedZone("", 2*3600))
		wantUtc, _ := time.Parse(time.RFC3339, "2024-06-23T13:00:00Z")
		if !race.StartDate.Equal(wantLocal) {
			t.Errorf("race StartDate = %v, want %v", race.StartDate, wantLocal)
		}
		if !race.StartDateUtc.Equal(wantUtc) {
			t.Errorf("race StartDateUtc = %v, want %v", race.StartDateUtc, wantUtc)
		}

		if race.StartDate.Hour() != 15 || race.StartDateUtc.Hour() != 13 {
			t.Errorf("local/UTC hour mismatch: local=%d utc=%d", race.StartDate.Hour(), race.StartDateUtc.Hour())
		}
		if race.Path != "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/" {
			t.Errorf("race Path = %q", race.Path)
		}
	}
	if !found {
		t.Fatal("expected to find the Spanish Grand Prix in the real fixture")
	}
}

func TestScheduleFromLivetiming_RealCapture_SprintWeekend(t *testing.T) {
	season := loadRealSeason(t)
	schedules := scheduleFromLivetiming(season)

	var found bool
	for _, s := range schedules {
		if s.EventName != "Austrian Grand Prix" {
			continue
		}
		found = true
		if s.EventFormat != "sprint" {
			t.Errorf("EventFormat = %q, want sprint", s.EventFormat)
		}
		wantTypes := []string{"Practice 1", "Sprint Qualifying", "Sprint", "Qualifying", "Race"}
		if len(s.Sessions) != len(wantTypes) {
			t.Fatalf("got %d sessions, want %d: %+v", len(s.Sessions), len(wantTypes), s.Sessions)
		}
		for i, want := range wantTypes {
			if s.Sessions[i].SessionType != want {
				t.Errorf("session %d = %q, want %q", i, s.Sessions[i].SessionType, want)
			}
		}
	}
	if !found {
		t.Fatal("expected to find the Austrian Grand Prix (a real 2024 sprint weekend) in the fixture")
	}
}

func TestDeriveEventFormat(t *testing.T) {
	cases := []struct {
		name     string
		event    string
		sessions []livetiming.Session
		want     string
	}{
		{"conventional", "Spanish Grand Prix", []livetiming.Session{{Name: "Practice 1"}, {Name: "Qualifying"}, {Name: "Race"}}, "conventional"},
		{"sprint", "Austrian Grand Prix", []livetiming.Session{{Name: "Practice 1"}, {Name: "Sprint Qualifying"}, {Name: "Sprint"}, {Name: "Qualifying"}, {Name: "Race"}}, "sprint"},
		{"testing", "Pre-Season Testing", []livetiming.Session{{Name: "Day 1"}}, "testing"},
		{"test", "Pre-Season Test", []livetiming.Session{{Name: "Day 1"}}, "testing"},
	}
	for _, c := range cases {
		if got := deriveEventFormat(c.event, c.sessions); got != c.want {
			t.Errorf("%s: deriveEventFormat = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestParseSessionTimes(t *testing.T) {
	local, utc, err := parseSessionTimes("2024-06-23T15:00:00", "02:00:00")
	if err != nil {
		t.Fatalf("parseSessionTimes: %v", err)
	}
	wantUtc, _ := time.Parse(time.RFC3339, "2024-06-23T13:00:00Z")
	if !utc.Equal(wantUtc) {
		t.Errorf("utc = %v, want %v", utc, wantUtc)
	}
	if local.Hour() != 15 {
		t.Errorf("local hour = %d, want 15", local.Hour())
	}
}

func TestParseSessionTimes_RejectsMalformed(t *testing.T) {
	if _, _, err := parseSessionTimes("not-a-date", "02:00:00"); err == nil {
		t.Fatal("expected an error for a malformed date")
	}
	if _, _, err := parseSessionTimes("2024-06-23T15:00:00", "bad"); err == nil {
		t.Fatal("expected an error for a malformed GmtOffset")
	}
}

func TestScheduleFromLivetiming_TestingEventsGetRoundZero(t *testing.T) {
	season := &livetiming.Season{Year: 2022, Meetings: []livetiming.Meeting{
		{Number: 2, Name: "Pre-Season Test"},
		{Number: 2, Name: "Saudi Arabian Grand Prix"},
	}}
	schedules := scheduleFromLivetiming(season)

	if got := schedules[0].RoundNumber; got != 0 {
		t.Errorf("Pre-Season Test RoundNumber = %d, want 0", got)
	}
	if got := schedules[1].RoundNumber; got != 2 {
		t.Errorf("Saudi Arabian Grand Prix RoundNumber = %d, want 2", got)
	}
}

func sched(name string, round int, format string, raceUtc string) jolpica.Schedule {
	race, _ := time.Parse(time.RFC3339, raceUtc)
	return jolpica.Schedule{Season: 2026, RoundNumber: round, EventName: name, EventFormat: format,
		Sessions: []jolpica.ScheduleSession{{SessionType: "Race", StartDateUtc: race}}}
}

func TestMerge_TakesJolpicaRoundsAndAddsMissingMeetings(t *testing.T) {
	lt := []jolpica.Schedule{
		sched("Pre-Season Testing", 0, "testing", "2026-02-13T07:00:00Z"),
		sched("Miami Grand Prix", 6, "sprint", "2026-05-03T20:00:00Z"),
		sched("Monaco Grand Prix", 6, "conventional", "2026-06-07T13:00:00Z"),
		sched("Bahrain Grand Prix", 16, "conventional", "2026-10-04T07:00:00Z"),
	}
	jp := []jolpica.Schedule{
		sched("Australian Grand Prix", 1, "", "2026-03-08T04:00:00Z"),
		sched("Miami Grand Prix", 4, "", "2026-05-03T20:00:00Z"),
		sched("Monaco Grand Prix", 6, "", "2026-06-07T13:00:00Z"),
		sched("Bahrain Grand Prix in Malaysia", 16, "", "2026-10-04T07:00:00Z"),
	}

	got := merge(lt, jp)

	want := []struct {
		name  string
		round int
	}{
		{"Pre-Season Testing", 0},
		{"Australian Grand Prix", 1},
		{"Miami Grand Prix", 4},
		{"Monaco Grand Prix", 6},
		{"Bahrain Grand Prix", 16},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i].EventName != w.name || got[i].RoundNumber != w.round {
			t.Errorf("event %d = %q round %d, want %q round %d", i, got[i].EventName, got[i].RoundNumber, w.name, w.round)
		}
	}

	aus := got[1]
	if len(aus.Sessions[0].PathCandidates) == 0 || aus.Sessions[0].PathCandidates[0] != "2026/2026-03-08_Australian_Grand_Prix/2026-03-08_Race/" {
		t.Errorf("Australian GP race candidates = %v", aus.Sessions[0].PathCandidates)
	}
}

func TestDerivePaths_TriesShiftedDates(t *testing.T) {
	race, _ := time.Parse(time.RFC3339, "2023-11-19T06:00:00Z")
	paths := derivePaths(2023, "Las Vegas Grand Prix", race, "Race", race)
	const want = "2023/2023-11-18_Las_Vegas_Grand_Prix/2023-11-18_Race/"
	found := false
	for _, p := range paths {
		if p == want {
			found = true
		}
	}
	if !found {
		t.Errorf("derivePaths did not include %q: %v", want, paths)
	}
	if paths[0] != "2023/2023-11-19_Las_Vegas_Grand_Prix/2023-11-19_Race/" {
		t.Errorf("first candidate = %q, want the unshifted path", paths[0])
	}
}
