package feeds

import (
	"os"
	"testing"
	"time"
)

func TestFixTyreStints_RealCapture_NoSpuriousChangesWithinSharedWindow(t *testing.T) {
	appData, err := os.ReadFile("testdata/timingappdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	timingData, err := os.ReadFile("testdata/timingdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	stints, err := ParseTyreStints(appData)
	if err != nil {
		t.Fatalf("ParseTyreStints: %v", err)
	}
	laps, err := ParseLapData(timingData)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}

	before := make(map[string][]TyreStint, len(stints))
	for drv, ls := range stints {
		cp := make([]TyreStint, len(ls))
		copy(cp, ls)
		before[drv] = cp
	}

	fixed := FixTyreStints(stints, laps, 0, false, true)
	_ = fixed

	sharedWindow := 59*time.Minute + 50*time.Second
	for drv, ls := range stints {
		for i, s := range ls {
			if before[drv][i].Timestamp >= sharedWindow {
				continue
			}
			if s != before[drv][i] {
				t.Errorf("driver %s entry %d changed unexpectedly within shared window: %+v -> %+v", drv, i, before[drv][i], s)
			}
		}
	}
}

func TestFixTyreStints_BunchedMessagesAtSessionStart(t *testing.T) {

	sessionStart := 10 * time.Second
	pitIn := 20 * time.Minute

	stints := map[string][]TyreStint{
		"1": {
			{DriverNumber: "1", Compound: "SOFT", Stint: 1, Timestamp: 15 * time.Second},
			{DriverNumber: "1", Compound: "MEDIUM", Stint: 2, Timestamp: 15 * time.Second},
		},
	}
	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 20, PitInTime: durPtr(pitIn)},
		},
	}

	FixTyreStints(stints, laps, sessionStart, true, true)

	s1, s2 := stints["1"][0], stints["1"][1]
	if s1.Timestamp != 1*time.Millisecond {
		t.Errorf("stint 1 (raw 0) Timestamp = %v, want %v (bracket start + 1ms)", s1.Timestamp, time.Millisecond)
	}
	want2 := pitIn + time.Millisecond
	if s2.Timestamp != want2 {
		t.Errorf("stint 2 (raw 1) Timestamp = %v, want %v", s2.Timestamp, want2)
	}
}

func TestFixTyreStints_NotAppliedInNonRaceLikeSessions(t *testing.T) {

	sessionStart := 10 * time.Second
	pitIn := 20 * time.Minute

	stints := map[string][]TyreStint{
		"1": {
			{DriverNumber: "1", Compound: "SOFT", Stint: 1, Timestamp: 15 * time.Second},
			{DriverNumber: "1", Compound: "MEDIUM", Stint: 2, Timestamp: 15 * time.Second},
		},
	}
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 20, PitInTime: durPtr(pitIn)}},
	}

	FixTyreStints(stints, laps, sessionStart, true, false)

	if stints["1"][0].Timestamp != 15*time.Second || stints["1"][1].Timestamp != 15*time.Second {
		t.Errorf("non-race-like session must not apply the bunched-message fix, got %+v", stints["1"])
	}
}

func TestFixTyreStints_StintCounterOvershootCorrected(t *testing.T) {

	pitIn := 20 * time.Minute

	stints := map[string][]TyreStint{
		"7": {
			{DriverNumber: "7", Compound: "SOFT", Stint: 1, Timestamp: 1 * time.Minute},
			{DriverNumber: "7", Compound: "SOFT", Stint: 2, Timestamp: 5 * time.Minute},
			{DriverNumber: "7", Compound: "MEDIUM", Stint: 2, Timestamp: 25 * time.Minute},
		},
	}
	laps := map[string][]LapData{
		"7": {{DriverNumber: "7", LapNo: 10, PitInTime: durPtr(pitIn)}},
	}

	fixed := FixTyreStints(stints, laps, 0, false, true)

	if len(fixed) != 1 || fixed[0] != "7" {
		t.Errorf("expected driver 7 reported as fixed, got %v", fixed)
	}
	if stints["7"][1].Stint != 1 {
		t.Errorf("overshooting message Stint = %d, want 1 (forced back to current bracket)", stints["7"][1].Stint)
	}
	if stints["7"][2].Stint != 2 {
		t.Errorf("genuine stint-2 message must be left alone, got Stint = %d", stints["7"][2].Stint)
	}
}

func TestFixTyreStints_NoStintDataForDriverInLaps(t *testing.T) {

	stints := map[string][]TyreStint{
		"99": {{DriverNumber: "99", Compound: "SOFT", Stint: 1, Timestamp: time.Second}},
	}
	fixed := FixTyreStints(stints, map[string][]LapData{}, 0, false, true)
	if len(fixed) != 0 {
		t.Errorf("expected no correction needed, got %v", fixed)
	}
}
