package feeds

import (
	"os"
	"testing"
)

func TestCalculatePositions_RealCapture_RanksByTimestamp(t *testing.T) {
	data, err := os.ReadFile("testdata/timingdata.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	laps, err := ParseLapData(data)
	if err != nil {
		t.Fatalf("ParseLapData: %v", err)
	}

	CalculatePositions(laps, true)

	byLapNo := make(map[int][]LapData)
	for _, ls := range laps {
		for _, l := range ls {
			byLapNo[l.LapNo] = append(byLapNo[l.LapNo], l)
		}
	}

	for lapNo, group := range byLapNo {
		seen := make(map[int]bool)
		for _, l := range group {
			if l.Position == nil {
				t.Errorf("lap %d driver %s: Position is nil, want assigned (race-like session)", lapNo, l.DriverNumber)
				continue
			}
			if *l.Position < 1 || *l.Position > len(group) {
				t.Errorf("lap %d driver %s: Position = %d out of range [1,%d]", lapNo, l.DriverNumber, *l.Position, len(group))
			}
			if seen[*l.Position] {
				t.Errorf("lap %d: duplicate Position %d", lapNo, *l.Position)
			}
			seen[*l.Position] = true
		}

		for i := range group {
			for j := range group {
				if *group[i].Position < *group[j].Position && group[i].Timestamp > group[j].Timestamp {
					t.Errorf("lap %d: driver %s ranked ahead of %s despite a later Timestamp", lapNo, group[i].DriverNumber, group[j].DriverNumber)
				}
			}
		}
	}
}

func TestCalculatePositions_NotAppliedInNonRaceLikeSessions(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 1, Timestamp: 10}},
	}
	CalculatePositions(laps, false)
	if laps["1"][0].Position != nil {
		t.Error("expected Position to remain nil for a non-race-like session")
	}
}

func TestCalculatePositions_GeneratedDnfLapExcluded(t *testing.T) {
	laps := map[string][]LapData{
		"1": {{DriverNumber: "1", LapNo: 1, Timestamp: 100, FastF1Generated: true}},
		"2": {{DriverNumber: "2", LapNo: 1, Timestamp: 50}},
	}
	CalculatePositions(laps, true)

	if laps["1"][0].Position != nil {
		t.Errorf("generated DNF lap Position = %v, want nil", laps["1"][0].Position)
	}
	if laps["2"][0].Position == nil || *laps["2"][0].Position != 1 {
		t.Errorf("driver 2 Position = %v, want 1 (sole real lap 1)", laps["2"][0].Position)
	}
}
