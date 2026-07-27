package feeds

import (
	"os"
	"testing"
	"time"
)

func TestAddTrackStatusToLaps_RealCapture_TailStatusApplies(t *testing.T) {
	timingData, err := os.ReadFile("testdata/timingdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	trackStatusData, err := os.ReadFile("testdata/trackstatus.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	laps, err := ParseLapData(timingData)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	statuses, err := ParseTrackStatus(trackStatusData)
	if err != nil {
		t.Fatalf("ParseTrackStatus: %v", err)
	}
	if len(statuses) != 2 {
		t.Fatalf("got %d track status events, want 2 (real fixture)", len(statuses))
	}

	AddTrackStatusToLaps(laps, statuses)

	drv4 := laps["4"]
	for i, l := range drv4 {
		if l.TrackStatus != "1" {
			t.Errorf("driver 4 lap %d: TrackStatus = %q, want %q", i, l.TrackStatus, "1")
		}
	}
}

func TestAddTrackStatusToLap_CaseA_LapEndsDuringWindow(t *testing.T) {
	statuses := []TrackStatus{
		{Status: "1", Timestamp: 0},
		{Status: "2", Timestamp: 10 * time.Minute},
		{Status: "1", Timestamp: 20 * time.Minute},
	}
	l := LapData{Timestamp: 5 * time.Minute}
	AddTrackStatusToLap(&l, statuses)
	if l.TrackStatus != "1" {
		t.Errorf("TrackStatus = %q, want %q", l.TrackStatus, "1")
	}
}

func TestAddTrackStatusToLap_CaseC_LapSpansMultipleWindows(t *testing.T) {

	statuses := []TrackStatus{
		{Status: "1", Timestamp: 0},
		{Status: "2", Timestamp: 10 * time.Minute},
		{Status: "1", Timestamp: 20 * time.Minute},
	}
	l := LapData{
		LapStartTime: durPtr(5 * time.Minute),
		Timestamp:    25 * time.Minute,
	}
	AddTrackStatusToLap(&l, statuses)
	if l.TrackStatus != "12" {
		t.Errorf("TrackStatus = %q, want %q (spans window 1 at start, window 2 fully, and the tail status 1 again without duplicating)", l.TrackStatus, "12")
	}
}

func TestAddTrackStatusToLaps_NoStatusesIsNoop(t *testing.T) {
	laps := map[string][]LapData{"1": {{Timestamp: time.Minute}}}
	AddTrackStatusToLaps(laps, nil)
	if laps["1"][0].TrackStatus != "" {
		t.Errorf("TrackStatus = %q, want empty (no-op with no status data)", laps["1"][0].TrackStatus)
	}
}
