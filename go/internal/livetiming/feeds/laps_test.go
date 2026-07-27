package feeds

import (
	"strconv"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseLapData_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingdata.jsonstream")

	result, err := ParseLapData(data)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}

	if len(result) != 20 {
		t.Fatalf("got %d drivers, want 20", len(result))
	}

	cases := []struct {
		driver string
		laps   []LapData
	}{
		{
			driver: "4",
			laps: []LapData{
				{
					LapNo: 1, Timestamp: 58*time.Minute + 21*time.Second + 509*time.Millisecond,
					LapTime:     nil,
					Sector1Time: nil,
					Sector2Time: durPtr(32*time.Second + 415*time.Millisecond),
					Sector3Time: durPtr(23*time.Second + 736*time.Millisecond),
					SpeedI1:     intPtr(254), SpeedI2: intPtr(260), SpeedFL: intPtr(275), SpeedST: intPtr(270),
					PitOutTime: durPtr(20*time.Minute + 33*time.Second + 427*time.Millisecond),
					PitStopNo:  0,
				},
				{
					LapNo: 2, Timestamp: 59*time.Minute + 41*time.Second + 724*time.Millisecond,
					LapTime:     durPtr(1*time.Minute + 20*time.Second + 215*time.Millisecond),
					Sector1Time: durPtr(24*time.Second + 310*time.Millisecond),
					Sector2Time: durPtr(32*time.Second + 14*time.Millisecond),
					Sector3Time: durPtr(23*time.Second + 891*time.Millisecond),
					SpeedI1:     intPtr(261), SpeedI2: intPtr(264), SpeedFL: intPtr(276), SpeedST: intPtr(299),
					IsPersonalBest: true,
					PitStopNo:      0,
				},
			},
		},
	}

	for _, tc := range cases {
		laps, ok := result[tc.driver]
		if !ok {
			t.Fatalf("driver %s not found", tc.driver)
		}
		if len(laps) != len(tc.laps) {
			t.Fatalf("driver %s: got %d laps, want %d", tc.driver, len(laps), len(tc.laps))
		}
		for i, want := range tc.laps {
			got := laps[i]
			if got.LapNo != want.LapNo {
				t.Errorf("driver %s lap %d: LapNo = %d, want %d", tc.driver, i, got.LapNo, want.LapNo)
			}
			if got.Timestamp != want.Timestamp {
				t.Errorf("driver %s lap %d: Timestamp = %v, want %v", tc.driver, i, got.Timestamp, want.Timestamp)
			}
			if !durEqual(got.LapTime, want.LapTime) {
				t.Errorf("driver %s lap %d: LapTime = %s, want %s", tc.driver, i, fmtDurTest(got.LapTime), fmtDurTest(want.LapTime))
			}
			if !durEqual(got.Sector1Time, want.Sector1Time) {
				t.Errorf("driver %s lap %d: Sector1Time = %s, want %s", tc.driver, i, fmtDurTest(got.Sector1Time), fmtDurTest(want.Sector1Time))
			}
			if !durEqual(got.Sector2Time, want.Sector2Time) {
				t.Errorf("driver %s lap %d: Sector2Time = %s, want %s", tc.driver, i, fmtDurTest(got.Sector2Time), fmtDurTest(want.Sector2Time))
			}
			if !durEqual(got.Sector3Time, want.Sector3Time) {
				t.Errorf("driver %s lap %d: Sector3Time = %s, want %s", tc.driver, i, fmtDurTest(got.Sector3Time), fmtDurTest(want.Sector3Time))
			}
			if !intEqual(got.SpeedI1, want.SpeedI1) {
				t.Errorf("driver %s lap %d: SpeedI1 = %s, want %s", tc.driver, i, fmtIntTest(got.SpeedI1), fmtIntTest(want.SpeedI1))
			}
			if !intEqual(got.SpeedFL, want.SpeedFL) {
				t.Errorf("driver %s lap %d: SpeedFL = %s, want %s", tc.driver, i, fmtIntTest(got.SpeedFL), fmtIntTest(want.SpeedFL))
			}
			if got.IsPersonalBest != want.IsPersonalBest {
				t.Errorf("driver %s lap %d: IsPersonalBest = %v, want %v", tc.driver, i, got.IsPersonalBest, want.IsPersonalBest)
			}
			if got.PitStopNo != want.PitStopNo {
				t.Errorf("driver %s lap %d: PitStopNo = %d, want %d (must be offset-corrected from the raw formation-lap pit visits)", tc.driver, i, got.PitStopNo, want.PitStopNo)
			}
			if !durEqual(got.PitOutTime, want.PitOutTime) {
				t.Errorf("driver %s lap %d: PitOutTime = %s, want %s", tc.driver, i, fmtDurTest(got.PitOutTime), fmtDurTest(want.PitOutTime))
			}
		}
	}
}

func TestParseLapData_RejectsMalformed(t *testing.T) {
	if _, err := ParseLapData([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestSectorSumCrossCheck_DiscardsCorruptLapTime(t *testing.T) {
	laps := []*lapBuilder{{
		sector1: durPtr(30 * time.Second),
		sector2: durPtr(30 * time.Second),
		sector3: durPtr(30 * time.Second),
		lapTime: durPtr(80 * time.Second),
	}}
	errs := sectorSumCrossCheck(laps)
	if errs != 1 {
		t.Fatalf("got %d integrity errors, want 1", errs)
	}
	if laps[0].lapTime != nil {
		t.Error("corrupt LapTime should have been discarded")
	}
}

func TestSectorSumCrossCheck_ReconstructsExplicitNone(t *testing.T) {
	laps := []*lapBuilder{{
		sector1:             durPtr(25 * time.Second),
		sector2:             durPtr(30 * time.Second),
		sector3:             durPtr(25 * time.Second),
		lapTimeExplicitNone: true,
	}}
	sectorSumCrossCheck(laps)
	if laps[0].lapTime == nil {
		t.Fatal("expected LapTime to be reconstructed from the sector sum")
	}
	if *laps[0].lapTime != 80*time.Second {
		t.Errorf("got %v, want 80s", *laps[0].lapTime)
	}
}

func TestReconstructMissingValues_LapTimeFromSectorSum(t *testing.T) {
	laps := []*lapBuilder{
		{lapTime: durPtr(80 * time.Second)},
		{
			sector1: durPtr(25 * time.Second),
			sector2: durPtr(30 * time.Second),
			sector3: durPtr(25 * time.Second),
		},
	}
	reconstructMissingValues(laps)
	if laps[1].lapTime == nil {
		t.Fatal("expected LapTime to be reconstructed")
	}
	if *laps[1].lapTime != 80*time.Second {
		t.Errorf("got %v, want 80s", *laps[1].lapTime)
	}
}

func TestReconstructMissingValues_BackSolvesSingleMissingSector(t *testing.T) {
	laps := []*lapBuilder{
		{sector2: durPtr(30 * time.Second)},
		{
			sector1: durPtr(25 * time.Second),
			sector3: durPtr(25 * time.Second),
			lapTime: durPtr(80 * time.Second),
		},
	}
	reconstructMissingValues(laps)
	if laps[1].sector2 == nil {
		t.Fatal("expected sector2 to be back-solved")
	}
	if *laps[1].sector2 != 30*time.Second {
		t.Errorf("got %v, want 30s", *laps[1].sector2)
	}
}

func TestSyncLapEndTimes_PicksEarliestDerivation(t *testing.T) {
	laps := []*lapBuilder{
		{time: 100 * time.Second, timeSet: true},
		{
			time: 200 * time.Second, timeSet: true,
			sector3:        durPtr(20 * time.Second),
			sector3Session: durPtr(150 * time.Second),
		},
	}
	syncLapEndTimes(laps)
	if laps[1].time != 150*time.Second {
		t.Errorf("got %v, want 150s (the earlier, sector-derived candidate)", laps[1].time)
	}
}

func TestReconstructDeletedPersonalBests_HandlesReinstatement(t *testing.T) {

	laps := []*lapBuilder{
		{lapTime: durPtr(70 * time.Second)},
		{lapTime: durPtr(65 * time.Second)},
	}
	sightings := []pbEntry{
		{elapsed: 1 * time.Second, value: 70 * time.Second},
		{elapsed: 2 * time.Second, value: 65 * time.Second},
		{elapsed: 3 * time.Second, value: 70 * time.Second},
	}
	reconstructDeletedPersonalBests(laps, sightings)

	if !laps[0].isPersonalBest {
		t.Error("the 70s lap should be marked as having held the personal best")
	}
	if !laps[1].isPersonalBest {
		t.Error("the 65s lap should be marked as having held the personal best")
	}
}

func TestCorrectPitStopOffset_SubtractsFirstLapCount(t *testing.T) {
	laps := []*lapBuilder{
		{numberOfPitStops: 2},
		{numberOfPitStops: 2},
		{numberOfPitStops: 3},
	}
	correctPitStopOffset(laps)
	want := []int{0, 0, 1}
	for i, w := range want {
		if laps[i].numberOfPitStops != w {
			t.Errorf("lap %d: numberOfPitStops = %d, want %d", i, laps[i].numberOfPitStops, w)
		}
	}
}

func TestCorrectFirstLapPitIn_DiscardsSpuriousEntry(t *testing.T) {
	laps := []*lapBuilder{{
		pitInTime:  durPtr(10 * time.Second),
		pitOutTime: durPtr(20 * time.Second),
	}}
	correctFirstLapPitIn(laps)
	if laps[0].pitInTime != nil {
		t.Error("expected the spurious PitInTime to be discarded")
	}
}

func TestCorrectFirstLapPitIn_KeepsGenuineEntry(t *testing.T) {
	laps := []*lapBuilder{{
		pitInTime:  durPtr(30 * time.Second),
		pitOutTime: durPtr(20 * time.Second),
	}}
	correctFirstLapPitIn(laps)
	if laps[0].pitInTime == nil {
		t.Error("expected the genuine PitInTime to be kept")
	}
}

func intPtr(n int) *int { return &n }

func durEqual(a, b *time.Duration) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
func intEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
func fmtDurTest(p *time.Duration) string {
	if p == nil {
		return "<nil>"
	}
	return p.String()
}
func fmtIntTest(p *int) string {
	if p == nil {
		return "<nil>"
	}
	return strconv.Itoa(*p)
}
