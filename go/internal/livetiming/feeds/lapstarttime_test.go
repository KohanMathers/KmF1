package feeds

import (
	"os"
	"testing"
	"time"
)

func TestDeriveLapStartTimes_RealCapture_BaseAssumption(t *testing.T) {
	timingData, err := os.ReadFile("testdata/timingdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	statusData, err := os.ReadFile("testdata/sessionstatus.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	laps, err := ParseLapData(timingData)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	events, err := ParseSessionStatus(statusData)
	if err != nil {
		t.Fatalf("ParseSessionStatus: %v", err)
	}
	sessionStart, ok := SessionStartTime(events)
	if !ok {
		t.Fatal("expected a Started event in the real fixture")
	}

	DeriveLapStartTimes(laps, sessionStart, true, events, true)

	for drv, ls := range laps {
		if len(ls) == 0 {
			continue
		}
		if ls[0].LapStartTime == nil {
			t.Errorf("driver %s lap 0: LapStartTime is nil, want session start", drv)
		} else if *ls[0].LapStartTime != sessionStart {
			t.Errorf("driver %s lap 0: LapStartTime = %v, want session start %v", drv, *ls[0].LapStartTime, sessionStart)
		}
		for i := 1; i < len(ls); i++ {
			if ls[i].LapStartTime == nil {
				t.Errorf("driver %s lap %d: LapStartTime is nil", drv, i)
				continue
			}
			if *ls[i].LapStartTime != ls[i-1].Timestamp {
				t.Errorf("driver %s lap %d: LapStartTime = %v, want previous lap's Timestamp %v", drv, i, *ls[i].LapStartTime, ls[i-1].Timestamp)
			}
		}
	}
}

func TestDeriveLapStartTimes_PracticeSession_NoLap0Assumption(t *testing.T) {
	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 1, Timestamp: time.Minute},
			{DriverNumber: "1", LapNo: 2, Timestamp: 2 * time.Minute},
		},
	}
	DeriveLapStartTimes(laps, 10*time.Second, true, nil, false)

	if laps["1"][0].LapStartTime != nil {
		t.Errorf("practice lap 0 LapStartTime = %v, want nil (no session-start assumption applies)", laps["1"][0].LapStartTime)
	}
	if laps["1"][1].LapStartTime == nil || *laps["1"][1].LapStartTime != time.Minute {
		t.Errorf("lap 1 LapStartTime = %v, want previous lap's Timestamp", laps["1"][1].LapStartTime)
	}
}

func TestDeriveLapStartTimes_RedFlagRestart_RaceLike(t *testing.T) {

	restartTime := 30 * time.Minute
	pitOut := 31 * time.Minute

	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 1, Timestamp: 5 * time.Minute},
			{DriverNumber: "1", LapNo: 2, Timestamp: 10 * time.Minute},
			{DriverNumber: "1", LapNo: 3, Timestamp: 40 * time.Minute, PitOutTime: durPtr(pitOut)},
		},
	}
	events := []SessionStatusEvent{
		{Status: "Started", Timestamp: 0},
		{Status: "Aborted", Timestamp: 15 * time.Minute},
		{Status: "Started", Timestamp: restartTime},
		{Status: "Finished", Timestamp: 60 * time.Minute},
	}

	DeriveLapStartTimes(laps, 0, true, events, true)

	if got := laps["1"][2].LapStartTime; got == nil || *got != restartTime {
		t.Errorf("restart lap LapStartTime = %v, want %v", got, restartTime)
	}
}

func TestDeriveLapStartTimes_RedFlagRestart_PreservesExistingLapTime(t *testing.T) {

	restartTime := 30 * time.Minute
	pitOut := 31 * time.Minute
	existingStart := 20 * time.Minute
	lapTime := 90 * time.Second

	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 1, Timestamp: 20 * time.Minute, LapStartTime: durPtr(existingStart), PitOutTime: durPtr(pitOut), LapTime: &lapTime},
		},
	}
	events := []SessionStatusEvent{
		{Status: "Aborted", Timestamp: 15 * time.Minute},
		{Status: "Started", Timestamp: restartTime},
	}

	DeriveLapStartTimes(laps, 0, false, events, true)

	if got := laps["1"][0].LapStartTime; got == nil || *got != existingStart {
		t.Errorf("LapStartTime = %v, want unchanged %v (LapTime already known)", got, existingStart)
	}
}

func TestDeriveLapStartTimes_RedFlagRestart_NonRaceLikeClearsInstead(t *testing.T) {
	restartTime := 30 * time.Minute
	pitOut := 31 * time.Minute

	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 1, Timestamp: 40 * time.Minute, PitOutTime: durPtr(pitOut)},
		},
	}
	events := []SessionStatusEvent{
		{Status: "Aborted", Timestamp: 15 * time.Minute},
		{Status: "Started", Timestamp: restartTime},
	}

	laps["1"][0].LapStartTime = durPtr(1 * time.Minute)
	applyRedFlagRestarts(laps, events, false)

	if laps["1"][0].LapStartTime != nil {
		t.Errorf("non-race-like restart lap LapStartTime = %v, want nil (no safe assumption, left for PitOutTime fallback)", laps["1"][0].LapStartTime)
	}
}

func TestDeriveLapStartTimes_FallsBackToPitOutTime(t *testing.T) {
	pitOut := 12 * time.Minute
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 5, Timestamp: 15 * time.Minute, PitOutTime: durPtr(pitOut)}},
	}

	DeriveLapStartTimes(laps, 0, false, nil, false)

	if got := laps["1"][0].LapStartTime; got == nil || *got != pitOut {
		t.Errorf("LapStartTime = %v, want PitOutTime fallback %v", got, pitOut)
	}
}

func TestDeriveLapStartTimes_NoRestartFoundSkipsDriver(t *testing.T) {

	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 1, Timestamp: 5 * time.Minute}},
	}
	events := []SessionStatusEvent{
		{Status: "Aborted", Timestamp: 15 * time.Minute},
		{Status: "Started", Timestamp: 30 * time.Minute},
	}
	DeriveLapStartTimes(laps, 0, false, events, true)

	if laps["1"][0].LapStartTime != nil {
		t.Errorf("expected LapStartTime to remain nil, got %v", laps["1"][0].LapStartTime)
	}
}
