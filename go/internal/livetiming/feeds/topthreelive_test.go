package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseTopThreeLive_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/topthree.jsonstream")

	result, err := ParseTopThreeLive(data)
	if err != nil {
		t.Fatalf("ParseTopThreeLive: %v", err)
	}

	if len(result[0]) != 3 {
		t.Fatalf("slot 0: got %d snapshots, want 3", len(result[0]))
	}
	if len(result[1]) != 3 {
		t.Fatalf("slot 1: got %d snapshots, want 3", len(result[1]))
	}
	if len(result[2]) != 2 {
		t.Fatalf("slot 2: got %d snapshots, want 2", len(result[2]))
	}

	initial := result[0][0]
	want := TopThreeLive{
		Position: 1, DriverNumber: "4", Abbreviation: "NOR",
		BroadcastName: "L NORRIS", FullName: "Lando NORRIS",
		TeamName: "McLaren", TeamColour: "FF8000",
		LapTime: nil, LapState: 80,
		DiffToAhead: "", DiffToLeader: "",
		OverallFastest: false, PersonalFastest: false,
		Timestamp: 8*time.Second + 720*time.Millisecond,
	}
	if initial != want {
		t.Errorf("slot 0 initial snapshot = %+v, want %+v", initial, want)
	}

	third := result[0][2]
	if third.LapState != 64 {
		t.Errorf("slot 0 third snapshot LapState = %d, want 64", third.LapState)
	}
	if third.DriverNumber != "4" || third.Abbreviation != "NOR" {
		t.Errorf("slot 0 third snapshot lost carried-forward driver identity: %+v", third)
	}
	if third.DiffToAhead != "LAP 1" || third.DiffToLeader != "LAP 1" {
		t.Errorf("slot 0 third snapshot lost carried-forward diff strings: %+v", third)
	}

	slot1Initial := result[1][0]
	if slot1Initial.DriverNumber != "1" || slot1Initial.Abbreviation != "VER" {
		t.Errorf("slot 1 initial = %+v, want driver 1/VER", slot1Initial)
	}
	slot1Last := result[1][2]
	if slot1Last.DriverNumber != "1" || slot1Last.LapState != 80 {
		t.Errorf("slot 1 last snapshot = %+v, want driver 1 carried forward with LapState 80", slot1Last)
	}
}

func TestParseTopThreeLive_RejectsMalformed(t *testing.T) {
	if _, err := ParseTopThreeLive([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestParseTopThreeLive_EmptyLapTimeIsNil(t *testing.T) {
	data := []byte(`00:00:01.000{"Lines":[{"Position":"1","LapTime":"1:20.500"}]}` + "\r\n" +
		`00:00:02.000{"Lines":{"0":{"LapTime":""}}}` + "\r\n")
	result, err := ParseTopThreeLive(data)
	if err != nil {
		t.Fatalf("ParseTopThreeLive: %v", err)
	}
	if len(result[0]) != 2 {
		t.Fatalf("got %d snapshots, want 2", len(result[0]))
	}
	if result[0][0].LapTime == nil || *result[0][0].LapTime != time.Minute+20*time.Second+500*time.Millisecond {
		t.Errorf("first LapTime = %v, want 1:20.500", result[0][0].LapTime)
	}
	if result[0][1].LapTime != nil {
		t.Errorf("second LapTime = %v, want nil (explicit empty)", result[0][1].LapTime)
	}
}
