package jolpica

import (
	"os"
	"testing"
	"time"
)

func TestParseSchedule_RealCapture_StandardWeekend(t *testing.T) {
	data, err := os.ReadFile("testdata/schedule_2024.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	schedules, err := ParseSchedule(data)
	if err != nil {
		t.Fatalf("ParseSchedule: %v", err)
	}
	if len(schedules) != 24 {
		t.Fatalf("got %d rounds, want 24", len(schedules))
	}

	var canadian *Schedule
	for i := range schedules {
		if schedules[i].RoundNumber == 9 {
			canadian = &schedules[i]
		}
	}
	if canadian == nil {
		t.Fatal("expected round 9 (Canadian GP)")
	}
	if canadian.Season != 2024 {
		t.Errorf("Season = %d, want 2024", canadian.Season)
	}
	if canadian.EventName != "Canadian Grand Prix" {
		t.Errorf("EventName = %q", canadian.EventName)
	}
	if canadian.Country != "Canada" || canadian.Location != "Montreal" {
		t.Errorf("Country/Location = %q/%q", canadian.Country, canadian.Location)
	}

	wantTypes := []string{"Practice 1", "Practice 2", "Practice 3", "Qualifying", "Race"}
	if len(canadian.Sessions) != len(wantTypes) {
		t.Fatalf("got %d sessions, want %d: %+v", len(canadian.Sessions), len(wantTypes), canadian.Sessions)
	}
	for i, want := range wantTypes {
		if canadian.Sessions[i].SessionType != want {
			t.Errorf("session %d = %q, want %q", i, canadian.Sessions[i].SessionType, want)
		}
		if canadian.Sessions[i].StartDate.IsZero() == false {
			t.Errorf("session %d StartDate = %v, want zero (Jolpica has no local venue time)", i, canadian.Sessions[i].StartDate)
		}
	}

	wantRaceUtc, _ := time.Parse(time.RFC3339, "2024-06-09T18:00:00Z")
	race := canadian.Sessions[len(canadian.Sessions)-1]
	if race.SessionType != "Race" || !race.StartDateUtc.Equal(wantRaceUtc) {
		t.Errorf("race session = %+v, want Race at %v", race, wantRaceUtc)
	}
}

func TestParseSchedule_RealCapture_SprintWeekendOrdering(t *testing.T) {
	data, err := os.ReadFile("testdata/schedule_2024.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	schedules, err := ParseSchedule(data)
	if err != nil {
		t.Fatalf("ParseSchedule: %v", err)
	}

	var chinese *Schedule
	for i := range schedules {
		if schedules[i].RoundNumber == 5 {
			chinese = &schedules[i]
		}
	}
	if chinese == nil {
		t.Fatal("expected round 5 (Chinese GP)")
	}

	wantTypes := []string{"Practice 1", "Sprint Qualifying", "Sprint", "Qualifying", "Race"}
	if len(chinese.Sessions) != len(wantTypes) {
		t.Fatalf("got %d sessions, want %d: %+v", len(chinese.Sessions), len(wantTypes), chinese.Sessions)
	}
	for i, want := range wantTypes {
		if chinese.Sessions[i].SessionType != want {
			t.Errorf("session %d = %q, want %q", i, chinese.Sessions[i].SessionType, want)
		}
	}
	sprint := chinese.Sessions[2]
	qualifying := chinese.Sessions[3]
	if !sprint.StartDateUtc.Before(qualifying.StartDateUtc) {
		t.Errorf("expected Sprint (%v) to be before Qualifying (%v)", sprint.StartDateUtc, qualifying.StartDateUtc)
	}
}

func TestParseSchedule_RejectsMalformed(t *testing.T) {
	if _, err := ParseSchedule([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
