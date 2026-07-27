package session

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestBackfillFirstLapTime_RealCapture(t *testing.T) {

	timingData := testfixture.Read(t, "../livetiming/feeds/testdata/timingdata.jsonstream")
	resultsData := testfixture.Read(t, "../jolpica/testdata/results_2024_9.json")
	lapsData := testfixture.Read(t, "../jolpica/testdata/laps1_2024_9.json")

	laps, err := feeds.ParseLapData(timingData)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}
	results, err := jolpica.ParseRaceResults(resultsData)
	if err != nil {
		t.Fatalf("ParseRaceResults: %v", err)
	}
	lap1Times, err := jolpica.ParseLapTimes(lapsData)
	if err != nil {
		t.Fatalf("ParseLapTimes: %v", err)
	}

	if laps["4"][0].LapNo != 1 || laps["4"][0].LapTime != nil {
		t.Fatalf("test fixture assumption changed: driver 4 lap 1 = %+v", laps["4"][0])
	}

	BackfillFirstLapTime(laps, results, lap1Times)

	want := time.Minute + 46*time.Second + 722*time.Millisecond
	if laps["4"][0].LapTime == nil || *laps["4"][0].LapTime != want {
		t.Errorf("driver 4 lap 1 LapTime = %v, want %v (Norris's real Jolpica lap-1 time)", laps["4"][0].LapTime, want)
	}

	if laps["4"][1].LapNo != 2 {
		t.Fatalf("test fixture assumption changed: laps[\"4\"][1] = %+v", laps["4"][1])
	}
	lap2Before := laps["4"][1].LapTime
	if lap2Before == nil {
		t.Fatal("expected lap 2 to already have a real LapTime before backfill")
	}
}

func TestBackfillFirstLapTime_DoesNotOverwriteExistingLapTime(t *testing.T) {
	existing := 90 * time.Second
	laps := map[string][]feeds.LapData{
		"1": {{DriverNumber: "1", LapNo: 1, LapTime: &existing}},
	}
	results := []jolpica.SessionResults{{DriverNumber: "1", DriverId: "driver_a"}}
	lap1Times := []jolpica.LapTiming{{LapNumber: 1, DriverId: "driver_a", Time: 80 * time.Second}}

	BackfillFirstLapTime(laps, results, lap1Times)

	if *laps["1"][0].LapTime != existing {
		t.Errorf("LapTime = %v, want unchanged %v", *laps["1"][0].LapTime, existing)
	}
}

func TestBackfillFirstLapTime_NoJoinFoundIsNoop(t *testing.T) {
	laps := map[string][]feeds.LapData{
		"1": {{DriverNumber: "1", LapNo: 1}},
	}

	results := []jolpica.SessionResults{{DriverNumber: "1"}}
	lap1Times := []jolpica.LapTiming{{LapNumber: 1, DriverId: "driver_a", Time: 80 * time.Second}}

	BackfillFirstLapTime(laps, results, lap1Times)

	if laps["1"][0].LapTime != nil {
		t.Errorf("LapTime = %v, want nil (no DriverId join available)", laps["1"][0].LapTime)
	}
}

func TestBackfillFirstLapTime_EmptyLapTimesIsNoop(t *testing.T) {
	laps := map[string][]feeds.LapData{
		"1": {{DriverNumber: "1", LapNo: 1}},
	}
	BackfillFirstLapTime(laps, nil, nil)
	if laps["1"][0].LapTime != nil {
		t.Error("expected no change with empty lap1Times")
	}
}
