package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestAlignLaps_CorrectsFollowerToMatchReportedGap(t *testing.T) {

	laps := map[string][]LapData{
		"L": {{Timestamp: 100 * time.Second}, {Timestamp: 200 * time.Second}},
		"F": {{Timestamp: 100*time.Second + 800*time.Millisecond}, {Timestamp: 201 * time.Second}},
	}
	intervals := map[string][]TimingInterval{
		"L": {{GapToLeader: "LAP 2", Timestamp: 100 * time.Second}},
		"F": {{GapToLeader: "+0.500", Timestamp: 100*time.Second + 800*time.Millisecond}},
	}

	unaligned := AlignLaps(laps, intervals)
	if len(unaligned) != 0 {
		t.Fatalf("got unaligned drivers %v, want none", unaligned)
	}

	if laps["L"][0].Timestamp != 100*time.Second {
		t.Errorf("L lap 0 = %v, want unchanged 100s", laps["L"][0].Timestamp)
	}
	if laps["L"][1].Timestamp != 200*time.Second {
		t.Errorf("L lap 1 = %v, want unchanged 200s", laps["L"][1].Timestamp)
	}

	wantLap0 := 100*time.Second + 500*time.Millisecond
	if laps["F"][0].Timestamp != wantLap0 {
		t.Errorf("F lap 0 = %v, want %v (must match the reported +0.500 gap exactly)", laps["F"][0].Timestamp, wantLap0)
	}
	gapAtLap0 := laps["F"][0].Timestamp - laps["L"][0].Timestamp
	if gapAtLap0 != 500*time.Millisecond {
		t.Errorf("corrected gap at lap 0 = %v, want 500ms", gapAtLap0)
	}
	wantLap1 := 201*time.Second - 300*time.Millisecond
	if laps["F"][1].Timestamp != wantLap1 {
		t.Errorf("F lap 1 = %v, want %v (same constant -0.3s shift applied to every lap)", laps["F"][1].Timestamp, wantLap1)
	}
}

func TestAlignLaps_ShiftsEverythingWhenDeltaIsPositive(t *testing.T) {

	laps := map[string][]LapData{
		"L": {{Timestamp: 100 * time.Second}},
		"F": {{Timestamp: 100*time.Second + 800*time.Millisecond}},
	}
	intervals := map[string][]TimingInterval{
		"L": {{GapToLeader: "LAP 1", Timestamp: 100 * time.Second}},
		"F": {{GapToLeader: "+1.000", Timestamp: 100*time.Second + 800*time.Millisecond}},
	}

	AlignLaps(laps, intervals)

	if laps["F"][0].Timestamp != 100*time.Second+800*time.Millisecond {
		t.Errorf("F lap 0 = %v, want unchanged", laps["F"][0].Timestamp)
	}
	wantL := 100*time.Second - 200*time.Millisecond
	if laps["L"][0].Timestamp != wantL {
		t.Errorf("L lap 0 = %v, want %v", laps["L"][0].Timestamp, wantL)
	}
	gap := laps["F"][0].Timestamp - laps["L"][0].Timestamp
	if gap != time.Second {
		t.Errorf("corrected gap = %v, want 1s (matching the reported +1.000 gap)", gap)
	}
}

func TestAlignLaps_LeavesUnresolvableDriverUntouchedAndReportsIt(t *testing.T) {
	laps := map[string][]LapData{
		"L": {{Timestamp: 100 * time.Second}},
		"X": {{Timestamp: 150 * time.Second}},
	}
	intervals := map[string][]TimingInterval{
		"L": {{GapToLeader: "LAP 1", Timestamp: 100 * time.Second}},
		"X": {{GapToLeader: "", Timestamp: 150 * time.Second}},
	}

	unaligned := AlignLaps(laps, intervals)

	if len(unaligned) != 1 || unaligned[0] != "X" {
		t.Errorf("got unaligned %v, want [\"X\"]", unaligned)
	}
	if laps["X"][0].Timestamp != 150*time.Second {
		t.Error("driver with no resolvable delta must be left completely unmodified")
	}
}

func TestAlignLaps_NoOpWhenNoGapDataAtAll(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{Timestamp: 100 * time.Second}},
	}
	intervals := map[string][]TimingInterval{
		"1": {{GapToLeader: "", Timestamp: 100 * time.Second}},
	}

	unaligned := AlignLaps(laps, intervals)
	if unaligned != nil {
		t.Errorf("got %v, want nil (bail out silently, matching fastf1's own early-return)", unaligned)
	}
	if laps["1"][0].Timestamp != 100*time.Second {
		t.Error("timestamps must be untouched when there's no gap data to align on")
	}
}

func TestAlignLaps_SkipsPittingDriverForLeaderDetection(t *testing.T) {

	pitTime := 50 * time.Second
	laps := map[string][]LapData{
		"L":   {{Timestamp: 100 * time.Second}},
		"Pit": {{Timestamp: 100 * time.Second, PitInTime: &pitTime}},
	}
	intervals := map[string][]TimingInterval{
		"L":   {{GapToLeader: "LAP 1", Timestamp: 100 * time.Second}},
		"Pit": {{GapToLeader: "+5.000", Timestamp: 100 * time.Second}},
	}

	unaligned := AlignLaps(laps, intervals)

	found := false
	for _, d := range unaligned {
		if d == "Pit" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected \"Pit\" to be unaligned (excluded from gap detection on its in/out lap), got %v", unaligned)
	}
}

func TestAlignLaps_RealCaptureSmokeTest(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingdata.jsonstream")

	laps, err := ParseLapData(data)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	intervals, err := ParseTimingIntervals(data)
	if err != nil {
		t.Fatalf("ParseTimingIntervals: %v", err)
	}

	before := laps["4"][1].Timestamp

	unaligned := AlignLaps(laps, intervals)
	t.Logf("unaligned drivers: %v", unaligned)

	after := laps["4"][1].Timestamp
	shift := after - before
	if shift < -5*time.Second || shift > 5*time.Second {
		t.Errorf("driver 4 lap 2 shifted by %v — implausibly large for a timing-jitter correction", shift)
	}
	t.Logf("driver 4 lap 2 shifted by %v (before=%v after=%v)", shift, before, after)
}
