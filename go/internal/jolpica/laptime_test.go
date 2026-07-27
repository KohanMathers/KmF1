package jolpica

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseLapTimes_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/laps1_2024_9.json")

	timings, err := ParseLapTimes(data)
	if err != nil {
		t.Fatalf("ParseLapTimes: %v", err)
	}
	if len(timings) != 20 {
		t.Fatalf("got %d timings, want 20", len(timings))
	}

	first := timings[0]
	want := LapTiming{LapNumber: 1, DriverId: "russell", Time: time.Minute + 43*time.Second + 723*time.Millisecond, Position: 1}
	if first != want {
		t.Errorf("first timing = %+v, want %+v", first, want)
	}

	last := timings[len(timings)-1]
	wantLast := LapTiming{LapNumber: 1, DriverId: "zhou", Time: 2*time.Minute + 3*time.Second + 563*time.Millisecond, Position: 20}
	if last != wantLast {
		t.Errorf("last timing = %+v, want %+v", last, wantLast)
	}
}

func TestParseLapTimes_RejectsMalformed(t *testing.T) {
	if _, err := ParseLapTimes([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
