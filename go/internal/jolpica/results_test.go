package jolpica

import (
	"os"
	"testing"
	"time"
)

func TestParseRaceResults_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/results_2024_9.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	results, err := ParseRaceResults(data)
	if err != nil {
		t.Fatalf("ParseRaceResults: %v", err)
	}
	if len(results) != 20 {
		t.Fatalf("got %d results, want 20", len(results))
	}

	leader := results[0]
	if leader.DriverNumber != "1" || leader.Abbreviation != "VER" || leader.TeamName != "Red Bull" {
		t.Errorf("leader identity = %+v", leader)
	}
	if leader.Position == nil || *leader.Position != 1 || leader.PositionText != "1" {
		t.Errorf("leader Position/PositionText = %v/%q", leader.Position, leader.PositionText)
	}
	if leader.GridPosition == nil || *leader.GridPosition != 2 {
		t.Errorf("leader GridPosition = %v, want 2", leader.GridPosition)
	}
	if leader.Laps == nil || *leader.Laps != 70 {
		t.Errorf("leader Laps = %v, want 70", leader.Laps)
	}
	if leader.Points != 25 {
		t.Errorf("leader Points = %v, want 25", leader.Points)
	}
	wantTime := time.Hour + 45*time.Minute + 47*time.Second + 927*time.Millisecond
	if leader.Time == nil || *leader.Time != wantTime {
		t.Errorf("leader Time (absolute total race time) = %v, want %v", leader.Time, wantTime)
	}
	if leader.FastestLapRank == nil || *leader.FastestLapRank != 4 {
		t.Errorf("leader FastestLapRank = %v, want 4", leader.FastestLapRank)
	}
	if leader.FastestLapNumber == nil || *leader.FastestLapNumber != 70 {
		t.Errorf("leader FastestLapNumber = %v, want 70", leader.FastestLapNumber)
	}
	wantFastest := 1*time.Minute + 15*time.Second + 569*time.Millisecond
	if leader.FastestLapTime == nil || *leader.FastestLapTime != wantFastest {
		t.Errorf("leader FastestLapTime = %v, want %v", leader.FastestLapTime, wantFastest)
	}
	if leader.FastestLapAvgSpeed == nil || *leader.FastestLapAvgSpeed != 207.751 {
		t.Errorf("leader FastestLapAvgSpeed = %v, want 207.751", leader.FastestLapAvgSpeed)
	}

	second := results[1]
	wantGap := 3*time.Second + 879*time.Millisecond
	if second.Time == nil || *second.Time != wantGap {
		t.Errorf("second-place Time (gap to leader, '+' stripped) = %v, want %v", second.Time, wantGap)
	}

	var sainz *SessionResults
	for i := range results {
		if results[i].Abbreviation == "SAI" {
			sainz = &results[i]
		}
	}
	if sainz == nil {
		t.Fatal("expected to find SAI in the real capture")
	}
	if sainz.Position == nil || *sainz.Position != 16 || sainz.PositionText != "R" {
		t.Errorf("retired driver Position/PositionText = %v/%q, want 16/R", sainz.Position, sainz.PositionText)
	}
	if sainz.Time != nil {
		t.Errorf("retired driver Time = %v, want nil", sainz.Time)
	}
	if sainz.Status != "Retired" {
		t.Errorf("retired driver Status = %q, want Retired", sainz.Status)
	}
}

func TestParseSprintResults_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/sprint_2024_6.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	results, err := ParseSprintResults(data)
	if err != nil {
		t.Fatalf("ParseSprintResults: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one sprint result")
	}

	leader := results[0]
	if leader.Abbreviation != "VER" || leader.Position == nil || *leader.Position != 1 {
		t.Errorf("sprint leader = %+v", leader)
	}
	wantTime := 31*time.Minute + 31*time.Second + 383*time.Millisecond
	if leader.Time == nil || *leader.Time != wantTime {
		t.Errorf("sprint leader Time = %v, want %v", leader.Time, wantTime)
	}
	if leader.Points != 8 {
		t.Errorf("sprint leader Points = %v, want 8", leader.Points)
	}
}

func TestParseQualifyingResults_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/qualifying_2024_9.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	results, err := ParseQualifyingResults(data)
	if err != nil {
		t.Fatalf("ParseQualifyingResults: %v", err)
	}
	if len(results) != 20 {
		t.Fatalf("got %d results, want 20", len(results))
	}

	pole := results[0]
	if pole.Abbreviation != "RUS" || pole.Position == nil || *pole.Position != 1 {
		t.Errorf("pole = %+v", pole)
	}
	wantQ1 := time.Minute + 13*time.Second + 13*time.Millisecond
	wantQ2 := time.Minute + 11*time.Second + 742*time.Millisecond
	wantQ3 := time.Minute + 12*time.Second
	if pole.Q1 == nil || *pole.Q1 != wantQ1 {
		t.Errorf("pole Q1 = %v, want %v", pole.Q1, wantQ1)
	}
	if pole.Q2 == nil || *pole.Q2 != wantQ2 {
		t.Errorf("pole Q2 = %v, want %v", pole.Q2, wantQ2)
	}
	if pole.Q3 == nil || *pole.Q3 != wantQ3 {
		t.Errorf("pole Q3 = %v, want %v", pole.Q3, wantQ3)
	}

	last := results[len(results)-1]
	if last.Abbreviation != "ZHO" {
		t.Errorf("last = %+v, want ZHO", last)
	}
	if last.Q1 == nil {
		t.Error("last driver Q1 = nil, want a real time")
	}
	if last.Q2 != nil || last.Q3 != nil {
		t.Errorf("last driver Q2/Q3 = %v/%v, want nil (knocked out in Q1)", last.Q2, last.Q3)
	}

	if last.Laps != nil || last.Points != 0 || last.Status != "" {
		t.Errorf("qualifying result has race-only fields populated: %+v", last)
	}
}

func TestParseRaceResults_RejectsMalformed(t *testing.T) {
	if _, err := ParseRaceResults([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestParseErgastDurationPtr(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
		nilv bool
	}{
		{"", 0, true},
		{"+3.879", 3*time.Second + 879*time.Millisecond, false},
		{"1:45:47.927", time.Hour + 45*time.Minute + 47*time.Second + 927*time.Millisecond, false},
		{"31:31.383", 31*time.Minute + 31*time.Second + 383*time.Millisecond, false},
	}
	for _, c := range cases {
		got := parseErgastDurationPtr(c.in)
		if c.nilv {
			if got != nil {
				t.Errorf("parseErgastDurationPtr(%q) = %v, want nil", c.in, *got)
			}
			continue
		}
		if got == nil || *got != c.want {
			t.Errorf("parseErgastDurationPtr(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}
