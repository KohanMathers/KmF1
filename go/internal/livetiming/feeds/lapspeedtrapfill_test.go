package feeds

import (
	"os"
	"testing"
)

func TestForwardFillSpeedTraps_RealCapture_FillsRealGapWithoutOverwritingRealValues(t *testing.T) {
	data, err := os.ReadFile("testdata/timingdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	laps, err := ParseLapData(data)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}

	before := make(map[string][]LapData, len(laps))
	for drv, ls := range laps {
		cp := make([]LapData, len(ls))
		copy(cp, ls)
		before[drv] = cp
	}

	if before["1"][1].SpeedI1 != nil {
		t.Fatal("test fixture assumption changed: driver 1 lap index 1 no longer has a real SpeedI1 gap")
	}

	ForwardFillSpeedTraps(laps)

	if laps["1"][1].SpeedI1 == nil || *laps["1"][1].SpeedI1 != *before["1"][0].SpeedI1 {
		t.Errorf("driver 1 lap index 1: SpeedI1 = %v, want carried forward from lap index 0 (%v)",
			laps["1"][1].SpeedI1, before["1"][0].SpeedI1)
	}

	for drv, ls := range laps {
		for i, l := range ls {
			b := before[drv][i]
			if b.SpeedI1 != nil && !intPtrEqual(l.SpeedI1, b.SpeedI1) {
				t.Errorf("driver %s lap %d: SpeedI1 changed from a real value %v to %v", drv, i, *b.SpeedI1, l.SpeedI1)
			}
			if b.SpeedI2 != nil && !intPtrEqual(l.SpeedI2, b.SpeedI2) {
				t.Errorf("driver %s lap %d: SpeedI2 changed from a real value %v to %v", drv, i, *b.SpeedI2, l.SpeedI2)
			}
			if b.SpeedFL != nil && !intPtrEqual(l.SpeedFL, b.SpeedFL) {
				t.Errorf("driver %s lap %d: SpeedFL changed from a real value %v to %v", drv, i, *b.SpeedFL, l.SpeedFL)
			}
			if b.SpeedST != nil && !intPtrEqual(l.SpeedST, b.SpeedST) {
				t.Errorf("driver %s lap %d: SpeedST changed from a real value %v to %v", drv, i, *b.SpeedST, l.SpeedST)
			}
		}
	}
}

func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func TestForwardFillSpeedTraps_FillsGapAcrossMissingValue(t *testing.T) {
	laps := map[string][]LapData{
		"1": {
			{LapNo: 1, SpeedI1: intPtr(300), SpeedI2: intPtr(290), SpeedFL: intPtr(310), SpeedST: intPtr(320)},
			{LapNo: 2},
			{LapNo: 3, SpeedI1: intPtr(301)},
		},
	}
	ForwardFillSpeedTraps(laps)

	l2 := laps["1"][1]
	if *l2.SpeedI1 != 300 || *l2.SpeedI2 != 290 || *l2.SpeedFL != 310 || *l2.SpeedST != 320 {
		t.Errorf("lap 2 = %+v, want all values carried forward from lap 1", l2)
	}
	l3 := laps["1"][2]
	if *l3.SpeedI1 != 301 {
		t.Errorf("lap 3 SpeedI1 = %v, want its own real value 301 (not overwritten)", *l3.SpeedI1)
	}
	if *l3.SpeedI2 != 290 {
		t.Errorf("lap 3 SpeedI2 = %v, want carried forward from lap 1 via lap 2's gap", *l3.SpeedI2)
	}
}

func TestForwardFillSpeedTraps_SkipsGeneratedAndRedFlagLaps(t *testing.T) {
	laps := map[string][]LapData{
		"1": {
			{LapNo: 1, SpeedI1: intPtr(300)},
			{LapNo: 2, FastF1Generated: true},
			{LapNo: 3, TrackStatus: "5"},
			{LapNo: 4},
		},
	}
	ForwardFillSpeedTraps(laps)

	if laps["1"][1].SpeedI1 != nil {
		t.Errorf("generated lap SpeedI1 = %v, want nil (excluded from write-back)", laps["1"][1].SpeedI1)
	}
	if laps["1"][2].SpeedI1 != nil {
		t.Errorf("red-flag lap SpeedI1 = %v, want nil (excluded from write-back)", laps["1"][2].SpeedI1)
	}
	if laps["1"][3].SpeedI1 == nil || *laps["1"][3].SpeedI1 != 300 {
		t.Errorf("lap 4 SpeedI1 = %v, want carried forward across the excluded laps to 300", laps["1"][3].SpeedI1)
	}
}

func TestForwardFillSpeedTraps_SpeedFLExcludesInlaps(t *testing.T) {
	pit := durPtr(0)
	laps := map[string][]LapData{
		"1": {
			{LapNo: 1, SpeedFL: intPtr(280)},
			{LapNo: 2, PitInTime: pit},
			{LapNo: 3},
		},
	}
	ForwardFillSpeedTraps(laps)

	if laps["1"][1].SpeedFL != nil {
		t.Errorf("inlap SpeedFL = %v, want nil (excluded)", laps["1"][1].SpeedFL)
	}
	if laps["1"][2].SpeedFL == nil || *laps["1"][2].SpeedFL != 280 {
		t.Errorf("lap 3 SpeedFL = %v, want 280", laps["1"][2].SpeedFL)
	}
}
