package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseDriverBestStats_RealCapture_InitialEmptySnapshotEmitsNothing(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingstats.jsonstream")
	result, err := ParseDriverBestStats(data)
	if err != nil {
		t.Fatalf("ParseDriverBestStats: %v", err)
	}

	for drv, snapshots := range result {
		if len(snapshots) == 0 {
			continue
		}
		if snapshots[0].Timestamp == 8*time.Second+738*time.Millisecond {
			t.Errorf("driver %s: got a snapshot from the all-empty initial record, want none", drv)
		}
	}
}

func TestParseDriverBestStats_RealCapture_PersonalBestLapTime(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingstats.jsonstream")
	result, err := ParseDriverBestStats(data)
	if err != nil {
		t.Fatalf("ParseDriverBestStats: %v", err)
	}

	drv63 := result["63"]
	var found *DriverBestStats
	for i, s := range drv63 {
		if s.Timestamp == 59*time.Minute+40*time.Second+24*time.Millisecond {
			found = &drv63[i]
		}
	}
	if found == nil {
		t.Fatal("expected a driver 63 snapshot at 00:59:40.024")
	}
	if found.BestLapNumber == nil || *found.BestLapNumber != 2 {
		t.Errorf("BestLapNumber = %v, want 2", found.BestLapNumber)
	}
	if found.BestLapPosition == nil || *found.BestLapPosition != 1 {
		t.Errorf("BestLapPosition = %v, want 1", found.BestLapPosition)
	}
	want := time.Minute + 19*time.Second + 857*time.Millisecond
	if found.BestLapTime == nil || *found.BestLapTime != want {
		t.Errorf("BestLapTime = %v, want %v", found.BestLapTime, want)
	}
}

func TestParseDriverBestStats_RealCapture_BestSpeedST(t *testing.T) {
	data := testfixture.Read(t, "testdata/timingstats.jsonstream")
	result, err := ParseDriverBestStats(data)
	if err != nil {
		t.Fatalf("ParseDriverBestStats: %v", err)
	}

	drv4 := result["4"]
	var found *DriverBestStats
	for i, s := range drv4 {
		if s.Timestamp == 57*time.Minute+7*time.Second+739*time.Millisecond {
			found = &drv4[i]
		}
	}
	if found == nil {
		t.Fatal("expected a driver 4 snapshot at 00:57:07.739")
	}
	if found.BestSpeedST == nil || *found.BestSpeedST != 270 {
		t.Errorf("BestSpeedST = %v, want 270", found.BestSpeedST)
	}
	if found.BestSpeedSTPosition == nil || *found.BestSpeedSTPosition != 1 {
		t.Errorf("BestSpeedSTPosition = %v, want 1", found.BestSpeedSTPosition)
	}
}

func TestParseDriverBestStats_BestSectorsArrayThenIndexedObject(t *testing.T) {

	data := []byte(
		`00:00:01.000{"Lines":{"1":{"BestSectors":[{"Position":1,"Value":"30.000"},{"Value":""},{"Value":""}]}}}` + "\r\n" +
			`00:00:02.000{"Lines":{"1":{"BestSectors":{"1":{"Position":1,"Value":"31.000"}}}}}` + "\r\n",
	)
	result, err := ParseDriverBestStats(data)
	if err != nil {
		t.Fatalf("ParseDriverBestStats: %v", err)
	}
	snapshots := result["1"]
	if len(snapshots) != 2 {
		t.Fatalf("got %d snapshots, want 2", len(snapshots))
	}
	if *snapshots[0].BestSector1Time != 30*time.Second {
		t.Errorf("snapshot 0 BestSector1Time = %v, want 30s (from array shape)", *snapshots[0].BestSector1Time)
	}
	if *snapshots[1].BestSector2Time != 31*time.Second {
		t.Errorf("snapshot 1 BestSector2Time = %v, want 31s (from indexed-object shape)", *snapshots[1].BestSector2Time)
	}

	if snapshots[1].BestSector1Time == nil || *snapshots[1].BestSector1Time != 30*time.Second {
		t.Errorf("snapshot 1 BestSector1Time = %v, want carried forward 30s", snapshots[1].BestSector1Time)
	}
}

func TestParseDriverBestStats_EmptyValueDoesNotEmitOrClobber(t *testing.T) {
	data := []byte(
		`00:00:01.000{"Lines":{"1":{"PersonalBestLapTime":{"Lap":1,"Position":1,"Value":"1:20.000"}}}}` + "\r\n" +
			`00:00:02.000{"Lines":{"1":{"PersonalBestLapTime":{"Value":""}}}}` + "\r\n",
	)
	result, err := ParseDriverBestStats(data)
	if err != nil {
		t.Fatalf("ParseDriverBestStats: %v", err)
	}

	if len(result["1"]) != 1 {
		t.Fatalf("got %d snapshots, want 1 (empty-value patch must not emit)", len(result["1"]))
	}
}

func TestParseDriverBestStats_RejectsMalformed(t *testing.T) {
	if _, err := ParseDriverBestStats([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
