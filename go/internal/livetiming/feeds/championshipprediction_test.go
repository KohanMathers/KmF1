package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseChampionshipPrediction_RealCapture_Drivers(t *testing.T) {
	data := testfixture.Read(t, "testdata/championshipprediction.jsonstream")

	drivers, _, err := ParseChampionshipPrediction(data)
	if err != nil {
		t.Fatalf("ParseChampionshipPrediction: %v", err)
	}

	drv1 := drivers["1"]
	if len(drv1) != 21 {
		t.Fatalf("driver 1: got %d snapshots, want 21", len(drv1))
	}

	first := drv1[0]
	want := ChampionshipPredictionDriver{
		DriverNumber: "1", CurrentPosition: 1, CurrentPoints: 194,
		PredictedPosition: 1, PredictedPoints: 194,
		Timestamp: 8*time.Second + 754*time.Millisecond,
	}
	if first != want {
		t.Errorf("driver 1 first snapshot = %+v, want %+v", first, want)
	}

	second := drv1[1]
	if second.CurrentPosition != 1 || second.CurrentPoints != 194 || second.PredictedPosition != 1 {
		t.Errorf("driver 1 second snapshot lost carried-forward fields: %+v", second)
	}
	if second.PredictedPoints != 212 {
		t.Errorf("driver 1 second snapshot PredictedPoints = %v, want 212", second.PredictedPoints)
	}
}

func TestParseChampionshipPrediction_RealCapture_Teams(t *testing.T) {
	data := testfixture.Read(t, "testdata/championshipprediction.jsonstream")

	_, teams, err := ParseChampionshipPrediction(data)
	if err != nil {
		t.Fatalf("ParseChampionshipPrediction: %v", err)
	}

	rbr := teams["Red Bull Racing Honda RBPT"]
	if len(rbr) == 0 {
		t.Fatal("expected snapshots for Red Bull Racing Honda RBPT")
	}
	first := rbr[0]
	want := ChampionshipPredictionTeam{
		TeamId: "Red Bull Racing Honda RBPT", TeamName: "Red Bull Racing",
		CurrentPosition: 1, CurrentPoints: 301,
		PredictedPosition: 1, PredictedPoints: 301,
		Timestamp: 8*time.Second + 754*time.Millisecond,
	}
	if first != want {
		t.Errorf("first team snapshot = %+v, want %+v", first, want)
	}

	if first.TeamId == first.TeamName {
		t.Error("TeamId and TeamName must be kept as separate, distinct fields")
	}
}

func TestParseChampionshipPrediction_EmptyInitialRecordEmitsNothing(t *testing.T) {
	data := []byte(`00:00:01.000{"Drivers":{},"Teams":{}}` + "\r\n")
	drivers, teams, err := ParseChampionshipPrediction(data)
	if err != nil {
		t.Fatalf("ParseChampionshipPrediction: %v", err)
	}
	if len(drivers) != 0 || len(teams) != 0 {
		t.Errorf("expected no snapshots from an all-empty record, got %d drivers / %d teams", len(drivers), len(teams))
	}
}

func TestParseChampionshipPrediction_RejectsMalformed(t *testing.T) {
	if _, _, err := ParseChampionshipPrediction([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
