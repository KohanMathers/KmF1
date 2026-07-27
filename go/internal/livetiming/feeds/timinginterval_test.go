package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseTimingIntervals_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingdata.jsonstream")

	result, err := ParseTimingIntervals(data)
	if err != nil {
		t.Fatalf("ParseTimingIntervals: %v", err)
	}

	drv4 := result["4"]
	if len(drv4) < 6 {
		t.Fatalf("got %d intervals for driver 4, want at least 6", len(drv4))
	}

	want := []TimingInterval{
		{DriverNumber: "4", Position: 1, GapToLeader: "", GapToCarAhead: "", Timestamp: 8*time.Second + 720*time.Millisecond},
		{DriverNumber: "4", Position: 1, GapToLeader: "LAP 1", GapToCarAhead: "LAP 1", Timestamp: time.Minute + 10*time.Second + 481*time.Millisecond},
		{DriverNumber: "4", Position: 2, GapToLeader: "", GapToCarAhead: "", Timestamp: 57*time.Minute + 9*time.Second + 799*time.Millisecond},
		{DriverNumber: "4", Position: 2, GapToLeader: "+0.037", GapToCarAhead: "+0.037", Timestamp: 57*time.Minute + 9*time.Second + 866*time.Millisecond},
		{DriverNumber: "4", Position: 3, GapToLeader: "+0.037", GapToCarAhead: "", Timestamp: 57*time.Minute + 14*time.Second + 962*time.Millisecond},
		{DriverNumber: "4", Position: 3, GapToLeader: "+0.781", GapToCarAhead: "+0.301", Timestamp: 57*time.Minute + 15*time.Second + 256*time.Millisecond},
	}

	for i, w := range want {
		got := drv4[i]
		if got != w {
			t.Errorf("interval %d = %+v, want %+v", i, got, w)
		}
	}
}

func TestParseTimingIntervals_OnlyEmitsOnChange(t *testing.T) {

	data := []byte(
		`00:00:01.000{"Lines":{"1":{"Position":"1","GapToLeader":"","IntervalToPositionAhead":{"Value":""}}}}` + "\r\n" +
			`00:00:02.000{"Lines":{"1":{"Position":"1"}}}` + "\r\n" +
			`00:00:03.000{"Lines":{"1":{"Position":"2"}}}` + "\r\n",
	)

	result, err := ParseTimingIntervals(data)
	if err != nil {
		t.Fatalf("ParseTimingIntervals: %v", err)
	}

	intervals := result["1"]
	if len(intervals) != 2 {
		t.Fatalf("got %d intervals, want 2 (the repeated Position:1 must not re-emit)", len(intervals))
	}
	if intervals[0].Position != 1 || intervals[1].Position != 2 {
		t.Errorf("got positions %d, %d, want 1, 2", intervals[0].Position, intervals[1].Position)
	}
}

func TestParseTimingIntervals_RejectsMalformed(t *testing.T) {
	if _, err := ParseTimingIntervals([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
