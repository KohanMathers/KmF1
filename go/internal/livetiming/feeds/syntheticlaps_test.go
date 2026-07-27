package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestGenerateSyntheticRetirementLaps_RealCapture_TruncatedWindowLooksLikeRetirement(t *testing.T) {

	timingData := testfixture.Read(t, "testdata/timingdata.jsonstream")
	statusData := testfixture.Read(t, "testdata/sessionstatus.jsonstream")
	trackStatusData := testfixture.Read(t, "testdata/trackstatus.jsonstream")
	carData, posData := readTelemetryFixtures(t)

	laps, err := ParseLapData(timingData)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	statusEvents, err := ParseSessionStatus(statusData)
	if err != nil {
		t.Fatalf("ParseSessionStatus: %v", err)
	}
	trackStatuses, err := ParseTrackStatus(trackStatusData)
	if err != nil {
		t.Fatalf("ParseTrackStatus: %v", err)
	}
	telemetry, err := ParseTelemetry(carData, posData)
	if err != nil {
		t.Fatalf("ParseTelemetry: %v", err)
	}

	before := make(map[string]int, len(laps))
	for drv, ls := range laps {
		before[drv] = len(ls)
	}

	GenerateSyntheticRetirementLaps(laps, statusEvents, trackStatuses, telemetry, 66, true)

	for drv, ls := range laps {
		if len(ls) != before[drv]+1 {
			t.Errorf("driver %s: length = %d, want %d (one synthetic lap appended)", drv, len(ls), before[drv]+1)
			continue
		}
		newLap := ls[len(ls)-1]
		if !newLap.FastF1Generated {
			t.Errorf("driver %s: new lap FastF1Generated = false, want true", drv)
		}
		if newLap.IsAccurate {
			t.Errorf("driver %s: new lap IsAccurate = true, want false (generated laps are never accurate)", drv)
		}
		if newLap.LapStartTime == nil || *newLap.LapStartTime != ls[len(ls)-2].Timestamp {
			t.Errorf("driver %s: new lap LapStartTime = %v, want previous lap's Timestamp %v", drv, newLap.LapStartTime, ls[len(ls)-2].Timestamp)
		}
		if newLap.Timestamp <= *newLap.LapStartTime {
			t.Errorf("driver %s: new lap Timestamp %v must be after its LapStartTime %v", drv, newLap.Timestamp, *newLap.LapStartTime)
		}
	}
}

func TestGenerateSyntheticRetirementLaps_SkipsAlreadyGeneratedSingleLap(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 1, Timestamp: time.Minute, FastF1Generated: true}},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: 2 * time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 0, false)
	if len(laps["1"]) != 1 {
		t.Errorf("expected no lap appended on top of an already-generated single lap, got %d laps", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_AlreadyCompletedFullDistance(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 66, Timestamp: time.Hour}},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: time.Hour + time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 1 {
		t.Errorf("expected no lap appended when LapNo already equals TotalLaps, got %d", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_AlreadyFinishedBeforeRefTime(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 50, Timestamp: time.Hour}},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: 30 * time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 1 {
		t.Errorf("expected no lap appended when a Finished status precedes refTime, got %d", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_InconclusiveNoFinishAfter(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 10, Timestamp: time.Hour}},
	}
	events := []SessionStatusEvent{{Status: "Aborted", Timestamp: time.Hour + time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 1 {
		t.Errorf("expected no lap appended when it's inconclusive whether the session ever finished, got %d", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_LastLapAlreadyAnInlap(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 10, Timestamp: time.Hour, PitInTime: durPtr(time.Hour - time.Second)}},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: time.Hour + time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 1 {
		t.Errorf("expected no lap appended when the last lap already has a PitInTime, got %d", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_RedundantRowCount(t *testing.T) {
	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 1, Timestamp: time.Minute},
			{DriverNumber: "1", LapNo: 2, Timestamp: 2 * time.Minute},
		},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: 3 * time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 2, true)
	if len(laps["1"]) != 2 {
		t.Errorf("expected no lap appended when row count already meets total laps, got %d", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_PitLaneDoubleEntryEdgeCase(t *testing.T) {
	laps := map[string][]LapData{
		"1": {
			{DriverNumber: "1", LapNo: 9, Timestamp: 50 * time.Minute, PitInTime: durPtr(49 * time.Minute)},
			{DriverNumber: "1", LapNo: 10, Timestamp: time.Hour},
		},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: time.Hour + time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 2 {
		t.Errorf("expected no synthetic lap for the pit-lane double-entry edge case, got %d laps", len(laps["1"]))
	}
}

func TestGenerateSyntheticRetirementLaps_AbortedBranchUsesRestartTimestamp(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 10, Timestamp: 50 * time.Minute}},
	}
	events := []SessionStatusEvent{
		{Status: "Aborted", Timestamp: 51 * time.Minute},
		{Status: "Finished", Timestamp: 90 * time.Minute},
	}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 2 {
		t.Fatalf("expected a synthetic lap to be appended, got %d laps", len(laps["1"]))
	}
	newLap := laps["1"][1]
	if newLap.Timestamp != 51*time.Minute {
		t.Errorf("Timestamp = %v, want the Aborted event's own timestamp %v", newLap.Timestamp, 51*time.Minute)
	}
}

func TestGenerateSyntheticRetirementLaps_ZeroSpeedTelemetryBranch(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 10, Timestamp: 50 * time.Minute}},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: 90 * time.Minute}}
	telemetry := map[string][]Telemetry{
		"1": {
			{DriverNumber: "1", Timestamp: 50*time.Minute + time.Second, Speed: intPtr(120)},
			{DriverNumber: "1", Timestamp: 50*time.Minute + 5*time.Second, Speed: intPtr(0)},
			{DriverNumber: "1", Timestamp: 50*time.Minute + 10*time.Second, Speed: intPtr(0)},
		},
	}
	GenerateSyntheticRetirementLaps(laps, events, nil, telemetry, 66, true)
	if len(laps["1"]) != 2 {
		t.Fatalf("expected a synthetic lap to be appended, got %d laps", len(laps["1"]))
	}
	want := 50*time.Minute + 5*time.Second
	if laps["1"][1].Timestamp != want {
		t.Errorf("Timestamp = %v, want first zero-speed sample %v", laps["1"][1].Timestamp, want)
	}
}

func TestGenerateSyntheticRetirementLaps_FallbackTo150Seconds(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 10, Timestamp: 50 * time.Minute}},
	}
	events := []SessionStatusEvent{{Status: "Finished", Timestamp: 90 * time.Minute}}
	GenerateSyntheticRetirementLaps(laps, events, nil, nil, 66, true)
	if len(laps["1"]) != 2 {
		t.Fatalf("expected a synthetic lap to be appended, got %d laps", len(laps["1"]))
	}
	want := 50*time.Minute + 150*time.Second
	if laps["1"][1].Timestamp != want {
		t.Errorf("Timestamp = %v, want the 150s flat fallback %v", laps["1"][1].Timestamp, want)
	}
}
