package session

import (
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/feeds"
	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestReconcileResults_RealCapture_MergesIdentityWithClassification(t *testing.T) {

	driverListData := testfixture.Read(t, "../livetiming/feeds/testdata/driverlist.jsonstream")
	resultsData := testfixture.Read(t, "../jolpica/testdata/results_2024_9.json")

	driverInfo, err := feeds.ParseDriverList(driverListData)
	if err != nil {
		t.Fatalf("ParseDriverList: %v", err)
	}
	jolpicaResults, err := jolpica.ParseRaceResults(resultsData)
	if err != nil {
		t.Fatalf("ParseRaceResults: %v", err)
	}

	merged := ReconcileResults(driverInfo, jolpicaResults, true)

	var ver *jolpica.SessionResults
	for i := range merged {
		if merged[i].DriverNumber == "1" {
			ver = &merged[i]
		}
	}
	if ver == nil {
		t.Fatal("expected driver 1 (VER) in the merged results")
	}
	if ver.Abbreviation != "VER" {
		t.Errorf("Abbreviation = %q, want VER (from the F1-API driver list)", ver.Abbreviation)
	}
	if ver.TeamName == "" {
		t.Error("TeamName is empty, want the F1-API driver list's team name")
	}
	if ver.Position == nil || *ver.Position != 1 {
		t.Errorf("Position = %v, want 1 (from Jolpica's classification)", ver.Position)
	}
	if ver.Points != 25 {
		t.Errorf("Points = %v, want 25 (from Jolpica)", ver.Points)
	}

	for i := 1; i < len(merged); i++ {
		pi, pj := merged[i-1].Position, merged[i].Position
		if pi != nil && pj != nil && *pi > *pj {
			t.Errorf("results not sorted by Position at index %d: %d before %d", i, *pi, *pj)
		}
	}
}

func TestReconcileResults_DriverInfoOnlyBuildsBareRows(t *testing.T) {
	driverInfo := map[string]*feeds.DriverInfo{
		"1": {DriverNumber: "1", Abbreviation: "VER", TeamName: "Red Bull Racing"},
	}
	merged := ReconcileResults(driverInfo, nil, true)
	if len(merged) != 1 {
		t.Fatalf("got %d rows, want 1", len(merged))
	}
	if merged[0].Abbreviation != "VER" || merged[0].Position != nil {
		t.Errorf("row = %+v, want VER with nil Position (no classification data at all)", merged[0])
	}
}

func TestReconcileResults_JolpicaOnlyUsesJolpicaDirectly(t *testing.T) {
	pos := 5
	jolpicaResults := []jolpica.SessionResults{
		{DriverNumber: "1", Abbreviation: "VER", Position: &pos},
	}
	merged := ReconcileResults(nil, jolpicaResults, true)
	if len(merged) != 1 || merged[0].Position == nil || *merged[0].Position != 5 {
		t.Errorf("merged = %+v, want the Jolpica row unchanged", merged)
	}
}

func TestReconcileResults_BothEmpty(t *testing.T) {
	if merged := ReconcileResults(nil, nil, true); merged != nil {
		t.Errorf("expected nil, got %+v", merged)
	}
}

func TestReconcileResults_OuterJoinAddsDNSRowWithNilPosition(t *testing.T) {

	pos, grid := 20, 18
	driverInfo := map[string]*feeds.DriverInfo{
		"1": {DriverNumber: "1", Abbreviation: "VER"},
	}
	jolpicaResults := []jolpica.SessionResults{
		{DriverNumber: "1", Abbreviation: "VER"},
		{DriverNumber: "99", Abbreviation: "DNS", Position: &pos, GridPosition: &grid},
	}

	merged := ReconcileResults(driverInfo, jolpicaResults, true)
	if len(merged) != 2 {
		t.Fatalf("got %d rows, want 2", len(merged))
	}

	var dns *jolpica.SessionResults
	for i := range merged {
		if merged[i].DriverNumber == "99" {
			dns = &merged[i]
		}
	}
	if dns == nil {
		t.Fatal("expected the outer-join DNS row for driver 99")
	}
	if dns.Position != nil || dns.GridPosition != nil {
		t.Errorf("DNS row = %+v, want Position and GridPosition forced nil", dns)
	}
}

func TestReconcileResults_LeftJoinExcludesJolpicaOnlyRowsWhenNotAccurate(t *testing.T) {

	driverInfo := map[string]*feeds.DriverInfo{
		"1": {DriverNumber: "1", Abbreviation: "VER"},
	}
	pos := 1
	jolpicaResults := []jolpica.SessionResults{
		{DriverNumber: "1", Abbreviation: "VER", Position: &pos},
		{DriverNumber: "99", Abbreviation: "DNS", Position: &pos},
	}

	merged := ReconcileResults(driverInfo, jolpicaResults, false)
	if len(merged) != 1 {
		t.Fatalf("got %d rows, want 1 (driver 99 excluded — left join only)", len(merged))
	}
	if merged[0].DriverNumber != "1" {
		t.Errorf("merged[0] = %+v, want driver 1", merged[0])
	}
}

func TestReconcileResults_FallsBackToJolpicaIdentityWhenF1APIMissing(t *testing.T) {

	driverInfo := map[string]*feeds.DriverInfo{
		"1": {DriverNumber: "1"},
	}
	jolpicaResults := []jolpica.SessionResults{
		{DriverNumber: "1", Abbreviation: "VER", TeamName: "Red Bull Racing"},
	}
	merged := ReconcileResults(driverInfo, jolpicaResults, true)
	if len(merged) != 1 || merged[0].Abbreviation != "VER" || merged[0].TeamName != "Red Bull Racing" {
		t.Errorf("merged = %+v, want Jolpica's identity used as fallback", merged)
	}
}
