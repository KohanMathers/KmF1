package session

import (
	"os"
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
)

func durPtr(d time.Duration) *time.Duration { return &d }

func TestQualifyingSplitTimes_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/qualifying_sessionstatus.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	events, err := feeds.ParseSessionStatus(data)
	if err != nil {
		t.Fatalf("ParseSessionStatus: %v", err)
	}

	splits := qualifyingSplitTimes(events)

	want := []time.Duration{
		13*time.Minute + 27*time.Second + 171*time.Millisecond,
		38*time.Minute + 27*time.Second + 192*time.Millisecond,
		time.Hour + 1*time.Minute + 27*time.Second + 193*time.Millisecond,
		time.Hour + 16*time.Minute + 47*time.Second + 412*time.Millisecond,
	}
	if len(splits) != len(want) {
		t.Fatalf("got %d splits, want %d: %v", len(splits), len(want), splits)
	}
	for i := range want {
		if splits[i] != want[i] {
			t.Errorf("split %d = %v, want %v", i, splits[i], want[i])
		}
	}
}

func TestQualifyingSplitTimes_RedFlagRestartSkipsExtraSplit(t *testing.T) {
	events := []feeds.SessionStatusEvent{
		{Status: "Started", Timestamp: 10 * time.Minute},
		{Status: "Aborted", Timestamp: 15 * time.Minute},
		{Status: "Started", Timestamp: 20 * time.Minute},
		{Status: "Finished", Timestamp: 30 * time.Minute},
		{Status: "Started", Timestamp: 35 * time.Minute},
		{Status: "Finished", Timestamp: 45 * time.Minute},
	}
	splits := qualifyingSplitTimes(events)
	want := []time.Duration{10 * time.Minute, 35 * time.Minute, 45 * time.Minute}
	if len(splits) != len(want) {
		t.Fatalf("got %d splits, want %d: %v", len(splits), len(want), splits)
	}
	for i := range want {
		if splits[i] != want[i] {
			t.Errorf("split %d = %v, want %v", i, splits[i], want[i])
		}
	}
}

func TestSplitQualifyingSessions_BasicWindowMembership(t *testing.T) {
	splitTimes := []time.Duration{0, 10 * time.Minute, 20 * time.Minute, 30 * time.Minute}
	laps := map[string][]feeds.LapData{
		"1": {
			{LapNo: 1, LapStartTime: durPtr(5 * time.Minute), Timestamp: 6 * time.Minute},
			{LapNo: 2, LapStartTime: durPtr(15 * time.Minute), Timestamp: 16 * time.Minute},
			{LapNo: 3, LapStartTime: durPtr(25 * time.Minute), Timestamp: 26 * time.Minute},
		},
	}
	buckets := splitQualifyingSessions(laps, splitTimes)
	if len(buckets[0]["1"]) != 1 || buckets[0]["1"][0].LapNo != 1 {
		t.Errorf("Q1 bucket = %+v, want just lap 1", buckets[0]["1"])
	}
	if len(buckets[1]["1"]) != 1 || buckets[1]["1"][0].LapNo != 2 {
		t.Errorf("Q2 bucket = %+v, want just lap 2", buckets[1]["1"])
	}
	if len(buckets[2]["1"]) != 1 || buckets[2]["1"][0].LapNo != 3 {
		t.Errorf("Q3 bucket = %+v, want just lap 3", buckets[2]["1"])
	}
}

func TestSplitQualifyingSessions_EarlyLapCarriesToNextWindow(t *testing.T) {
	splitTimes := []time.Duration{0, 10 * time.Minute, 20 * time.Minute}

	laps := map[string][]feeds.LapData{
		"1": {
			{LapNo: 1, LapStartTime: durPtr(9*time.Minute + 30*time.Second), Timestamp: 10*time.Minute + 30*time.Second, PitOutTime: durPtr(9 * time.Minute)},
		},
	}
	buckets := splitQualifyingSessions(laps, splitTimes)
	if len(buckets[0]["1"]) != 0 {
		t.Errorf("Q1 bucket = %+v, want empty (early lap excluded)", buckets[0]["1"])
	}
	if len(buckets[1]["1"]) != 1 {
		t.Errorf("Q2 bucket = %+v, want the carried-forward early lap", buckets[1]["1"])
	}
}

func TestPickQuicklaps_FiltersOutsideThreshold(t *testing.T) {
	bucket := map[string][]feeds.LapData{
		"1": {{LapTime: durPtr(80 * time.Second)}},
		"2": {{LapTime: durPtr(85 * time.Second)}},
		"3": {{LapTime: durPtr(90 * time.Second)}},
		"4": {{LapTime: nil}},
		"5": {{LapTime: durPtr(81 * time.Second), Deleted: true}},
	}
	kept := pickQuicklaps(bucket)
	if _, ok := kept["1"]; !ok {
		t.Error("driver 1 (fastest) should be kept")
	}
	if _, ok := kept["2"]; !ok {
		t.Error("driver 2 (within 107%) should be kept")
	}
	if _, ok := kept["3"]; ok {
		t.Error("driver 3 (outside 107%) should be dropped")
	}
	if _, ok := kept["4"]; ok {
		t.Error("driver 4 (no LapTime) should be dropped")
	}
	if _, ok := kept["5"]; ok {
		t.Error("driver 5 (deleted) should be dropped")
	}
}

func TestCalculateRaceResultsFromLaps_RanksByLapsThenTime(t *testing.T) {
	laps := map[string][]feeds.LapData{
		"1": {{LapNo: 1}, {LapNo: 2}, {LapNo: 3, Timestamp: 100 * time.Second}},
		"2": {{LapNo: 1}, {LapNo: 2}, {LapNo: 3, Timestamp: 90 * time.Second}},
		"3": {{LapNo: 1}, {LapNo: 2}},
	}
	results := []jolpica.SessionResults{
		{DriverNumber: "1"}, {DriverNumber: "2"}, {DriverNumber: "3"},
	}

	out := CalculateRaceResultsFromLaps(results, laps, false)

	byDriver := make(map[string]*jolpica.SessionResults, len(out))
	for i := range out {
		byDriver[out[i].DriverNumber] = &out[i]
	}
	if p := byDriver["2"].Position; p == nil || *p != 1 {
		t.Errorf("driver 2 Position = %v, want 1 (3 laps, finished first)", p)
	}
	if p := byDriver["1"].Position; p == nil || *p != 2 {
		t.Errorf("driver 1 Position = %v, want 2 (3 laps, finished second)", p)
	}
	if p := byDriver["3"].Position; p == nil || *p != 3 {
		t.Errorf("driver 3 Position = %v, want 3 (fewer laps)", p)
	}
}

func TestCalculateRaceResultsFromLaps_SkipsWhenPositionAlreadyKnown(t *testing.T) {
	pos := 1
	results := []jolpica.SessionResults{{DriverNumber: "1", Position: &pos}}
	laps := map[string][]feeds.LapData{"1": {{LapNo: 1}}}

	out := CalculateRaceResultsFromLaps(results, laps, false)
	if out[0].Position != &pos {
		t.Error("expected results to be returned unchanged when a Position is already known")
	}
}

func TestCalculateQualiResultsFromLaps_EndToEndSynthetic(t *testing.T) {
	statusEvents := []feeds.SessionStatusEvent{
		{Status: "Started", Timestamp: 0},
		{Status: "Finished", Timestamp: 20 * time.Minute},
		{Status: "Started", Timestamp: 20 * time.Minute},
		{Status: "Finished", Timestamp: 40 * time.Minute},
		{Status: "Started", Timestamp: 40 * time.Minute},
		{Status: "Finished", Timestamp: 60 * time.Minute},
	}

	mk := func(start time.Duration, lapTime time.Duration) feeds.LapData {
		return feeds.LapData{
			LapStartTime: durPtr(start), Timestamp: start + lapTime,
			LapTime: durPtr(lapTime), IsAccurate: true,
		}
	}

	laps := map[string][]feeds.LapData{

		"1": {mk(5*time.Minute, 90*time.Second), mk(25*time.Minute, 89*time.Second), mk(45*time.Minute, 87*time.Second)},

		"2": {mk(5*time.Minute, 91*time.Second)},
	}
	results := []jolpica.SessionResults{{DriverNumber: "1"}, {DriverNumber: "2"}}

	out := CalculateQualiResultsFromLaps(results, laps, statusEvents, false)

	byDriver := make(map[string]*jolpica.SessionResults, len(out))
	for i := range out {
		byDriver[out[i].DriverNumber] = &out[i]
	}

	d1, d2 := byDriver["1"], byDriver["2"]
	if d1.Position == nil || *d1.Position != 1 {
		t.Errorf("driver 1 Position = %v, want 1 (reached Q3)", d1.Position)
	}
	if d2.Position == nil || *d2.Position != 2 {
		t.Errorf("driver 2 Position = %v, want 2 (only reached Q1)", d2.Position)
	}
	if d1.Q1 == nil || *d1.Q1 != 90*time.Second {
		t.Errorf("driver 1 Q1 = %v, want 90s", d1.Q1)
	}
	if d1.Q3 == nil || *d1.Q3 != 87*time.Second {
		t.Errorf("driver 1 Q3 = %v, want 87s", d1.Q3)
	}
	if d2.Q2 != nil || d2.Q3 != nil {
		t.Errorf("driver 2 Q2/Q3 = %v/%v, want nil (knocked out in Q1)", d2.Q2, d2.Q3)
	}
}

func TestCalculateQualiResultsFromLaps_SkipsWhenPositionAlreadyKnown(t *testing.T) {
	pos := 1
	results := []jolpica.SessionResults{{DriverNumber: "1", Position: &pos}}
	out := CalculateQualiResultsFromLaps(results, map[string][]feeds.LapData{"1": {{}}}, []feeds.SessionStatusEvent{{Status: "Started"}}, false)
	if out[0].Position != &pos {
		t.Error("expected results to be returned unchanged when a Position is already known")
	}
}
