package jolpica

import (
	"os"
	"testing"
)

func TestParseDriverStandings_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/driverstandings_2024_9.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	standings, err := ParseDriverStandings(data)
	if err != nil {
		t.Fatalf("ParseDriverStandings: %v", err)
	}
	if len(standings) == 0 {
		t.Fatal("expected at least one standing")
	}

	leader := standings[0]
	want := DriverStanding{
		Season: 2024, Round: 9, Position: 1, PositionText: "1",
		Points: 194, Wins: 6, DriverId: "max_verstappen", Abbreviation: "VER",
		ConstructorIds: []string{"red_bull"},
	}
	if leader.Season != want.Season || leader.Round != want.Round || leader.Position != want.Position ||
		leader.PositionText != want.PositionText || leader.Points != want.Points || leader.Wins != want.Wins ||
		leader.DriverId != want.DriverId || leader.Abbreviation != want.Abbreviation {
		t.Errorf("leader = %+v, want %+v", leader, want)
	}
	if len(leader.ConstructorIds) != 1 || leader.ConstructorIds[0] != "red_bull" {
		t.Errorf("leader ConstructorIds = %v, want [red_bull]", leader.ConstructorIds)
	}
}

func TestParseConstructorStandings_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/constructorstandings_2024_9.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	standings, err := ParseConstructorStandings(data)
	if err != nil {
		t.Fatalf("ParseConstructorStandings: %v", err)
	}
	if len(standings) == 0 {
		t.Fatal("expected at least one standing")
	}

	leader := standings[0]
	if leader.Season != 2024 || leader.Round != 9 || leader.Position != 1 {
		t.Errorf("leader Season/Round/Position = %d/%d/%d", leader.Season, leader.Round, leader.Position)
	}
	if leader.ConstructorId != "red_bull" || leader.Points != 301 || leader.Wins != 6 {
		t.Errorf("leader = %+v", leader)
	}
}

func TestParseDriverStandings_RejectsMalformed(t *testing.T) {
	if _, err := ParseDriverStandings([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
