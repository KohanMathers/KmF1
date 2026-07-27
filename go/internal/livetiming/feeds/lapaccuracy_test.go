package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestCheckLapAccuracy_RealCapture(t *testing.T) {
	timingData := testfixture.Read(t, "testdata/timingdata.jsonstream")
	trackStatusData := testfixture.Read(t, "testdata/trackstatus.jsonstream")

	laps, err := ParseLapData(timingData)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	statuses, err := ParseTrackStatus(trackStatusData)
	if err != nil {
		t.Fatalf("ParseTrackStatus: %v", err)
	}
	AddTrackStatusToLaps(laps, statuses)

	CheckLapAccuracy(laps)

	drv4 := laps["4"]

	if drv4[0].IsAccurate {
		t.Error("driver 4 lap 1: IsAccurate = true, want false (no LapTime available)")
	}

	if !drv4[1].IsAccurate {
		t.Errorf("driver 4 lap 2: IsAccurate = false, want true (%+v)", drv4[1])
	}
}

func TestCheckLapAccuracy_FailsWithoutLapTime(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{TrackStatus: "1", Sector1Time: durPtr(time.Second), Sector2Time: durPtr(time.Second), Sector3Time: durPtr(time.Second)}},
	}
	CheckLapAccuracy(laps)
	if laps["1"][0].IsAccurate {
		t.Error("expected IsAccurate = false when LapTime is missing")
	}
}

func TestCheckLapAccuracy_FailsUnderSafetyCarOrPitStop(t *testing.T) {
	lapTime := 90 * time.Second
	third := 30 * time.Second
	cases := []struct {
		name string
		lap  LapData
	}{
		{"safety car status", LapData{TrackStatus: "4", LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
		{"red flag combo", LapData{TrackStatus: "45", LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
		{"has pit in", LapData{TrackStatus: "1", PitInTime: durPtr(time.Second), LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
		{"has pit out", LapData{TrackStatus: "1", PitOutTime: durPtr(time.Second), LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
		{"generated", LapData{TrackStatus: "1", FastF1Generated: true, LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
	}
	for _, c := range cases {
		laps := map[string][]LapData{"1": {c.lap}}
		CheckLapAccuracy(laps)
		if laps["1"][0].IsAccurate {
			t.Errorf("%s: IsAccurate = true, want false", c.name)
		}
	}
}

func TestCheckLapAccuracy_FailsAfterSafetyCarLap(t *testing.T) {
	lapTime := 90 * time.Second
	third := 30 * time.Second
	laps := map[string][]LapData{
		"1": {
			{TrackStatus: "4", Timestamp: 0, LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third},
			{TrackStatus: "1", Timestamp: 90 * time.Second, LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third},
		},
	}
	CheckLapAccuracy(laps)
	if laps["1"][1].IsAccurate {
		t.Error("lap immediately following a Safety Car lap: IsAccurate = true, want false")
	}
}

func TestCheckLapAccuracy_SectorSumMismatchIsIntegrityError(t *testing.T) {
	lapTime := 90 * time.Second
	third := 30 * time.Second
	wrongLapTime := 100 * time.Second
	laps := map[string][]LapData{
		"1": {{TrackStatus: "1", Timestamp: 0, LapTime: &wrongLapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
	}
	errs := CheckLapAccuracy(laps)
	if errs != 1 {
		t.Errorf("integrity errors = %d, want 1", errs)
	}
	if laps["1"][0].IsAccurate {
		t.Error("expected IsAccurate = false on sector-sum mismatch")
	}
	_ = lapTime
}

func TestCheckLapAccuracy_ToleratesSubMillisecondRoundingWithin3ms(t *testing.T) {
	lapTime := 90*time.Second + 2*time.Millisecond
	third := 30 * time.Second
	laps := map[string][]LapData{
		"1": {{TrackStatus: "1", Timestamp: 90 * time.Second, LapTime: &lapTime, Sector1Time: &third, Sector2Time: &third, Sector3Time: &third}},
	}
	errs := CheckLapAccuracy(laps)
	if errs != 0 {
		t.Errorf("integrity errors = %d, want 0 (within 3ms tolerance)", errs)
	}
	if !laps["1"][0].IsAccurate {
		t.Error("expected IsAccurate = true within 3ms sector-sum tolerance")
	}
}
