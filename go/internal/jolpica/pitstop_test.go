package jolpica

import (
	"os"
	"testing"
	"time"
)

func TestParsePitStops_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/pitstops_2024_9.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	stops, err := ParsePitStops(data)
	if err != nil {
		t.Fatalf("ParsePitStops: %v", err)
	}
	if len(stops) != 30 {
		t.Fatalf("got %d stops, want 30", len(stops))
	}

	first := stops[0]
	want := PitStop{
		DriverId: "kevin_magnussen", StopNumber: 1, Lap: 7,
		Duration: 30*time.Second + 472*time.Millisecond,
		Time:     14*time.Hour + 14*time.Minute + 25*time.Second,
	}
	if first != want {
		t.Errorf("first stop = %+v, want %+v", first, want)
	}

	foundStop2, foundStop3 := false, false
	for _, s := range stops {
		if s.StopNumber == 2 {
			foundStop2 = true
		}
		if s.StopNumber == 3 {
			foundStop3 = true
		}
	}
	if !foundStop2 || !foundStop3 {
		t.Errorf("expected to find real stop 2 and stop 3 entries, got stop2=%v stop3=%v", foundStop2, foundStop3)
	}
}

func TestParsePitStops_RejectsMalformed(t *testing.T) {
	if _, err := ParsePitStops([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestParseTimeOfDay(t *testing.T) {
	got, err := parseTimeOfDay("14:14:25")
	if err != nil {
		t.Fatalf("parseTimeOfDay: %v", err)
	}
	want := 14*time.Hour + 14*time.Minute + 25*time.Second
	if got != want {
		t.Errorf("parseTimeOfDay = %v, want %v", got, want)
	}
}
