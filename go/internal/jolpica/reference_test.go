package jolpica

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseDriverReferences_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/drivers_2024.json")

	drivers, err := ParseDriverReferences(data)
	if err != nil {
		t.Fatalf("ParseDriverReferences: %v", err)
	}
	if len(drivers) != 25 {
		t.Fatalf("got %d drivers, want 25", len(drivers))
	}

	var albon *DriverReference
	for i := range drivers {
		if drivers[i].DriverId == "albon" {
			albon = &drivers[i]
		}
	}
	if albon == nil {
		t.Fatal("expected to find driver 'albon'")
	}
	if albon.Abbreviation != "ALB" {
		t.Errorf("Abbreviation = %q, want ALB", albon.Abbreviation)
	}
	if albon.PermanentNumber == nil || *albon.PermanentNumber != 23 {
		t.Errorf("PermanentNumber = %v, want 23", albon.PermanentNumber)
	}
	if albon.GivenName != "Alexander" || albon.FamilyName != "Albon" {
		t.Errorf("name = %q %q", albon.GivenName, albon.FamilyName)
	}
	wantDob, _ := time.Parse("2006-01-02", "1996-03-23")
	if !albon.DateOfBirth.Equal(wantDob) {
		t.Errorf("DateOfBirth = %v, want %v", albon.DateOfBirth, wantDob)
	}
	if albon.Nationality != "Thai" {
		t.Errorf("Nationality = %q, want Thai", albon.Nationality)
	}
}

func TestParseConstructorReferences_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/constructors_2024.json")
	constructors, err := ParseConstructorReferences(data)
	if err != nil {
		t.Fatalf("ParseConstructorReferences: %v", err)
	}
	if len(constructors) != 10 {
		t.Fatalf("got %d constructors, want 10", len(constructors))
	}
	var ferrari *ConstructorReference
	for i := range constructors {
		if constructors[i].ConstructorId == "ferrari" {
			ferrari = &constructors[i]
		}
	}
	if ferrari == nil {
		t.Fatal("expected to find constructor 'ferrari'")
	}
	if ferrari.Name != "Ferrari" || ferrari.Nationality != "Italian" {
		t.Errorf("ferrari = %+v", ferrari)
	}
}

func TestParseCircuitReferences_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/circuits_2024.json")
	circuits, err := ParseCircuitReferences(data)
	if err != nil {
		t.Fatalf("ParseCircuitReferences: %v", err)
	}
	if len(circuits) != 24 {
		t.Fatalf("got %d circuits, want 24", len(circuits))
	}
	var catalunya *CircuitReference
	for i := range circuits {
		if circuits[i].CircuitId == "catalunya" {
			catalunya = &circuits[i]
		}
	}
	if catalunya == nil {
		t.Fatal("expected to find circuit 'catalunya'")
	}
	if catalunya.Name != "Circuit de Barcelona-Catalunya" {
		t.Errorf("Name = %q", catalunya.Name)
	}
	if catalunya.Locality != "Barcelona" || catalunya.Country != "Spain" {
		t.Errorf("Locality/Country = %q/%q", catalunya.Locality, catalunya.Country)
	}
	if catalunya.Latitude != 41.57 || catalunya.Longitude != 2.26111 {
		t.Errorf("Lat/Long = %v/%v, want 41.57/2.26111", catalunya.Latitude, catalunya.Longitude)
	}
}

func TestParseFinishingStatusReferences_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/status_2024.json")
	statuses, err := ParseFinishingStatusReferences(data)
	if err != nil {
		t.Fatalf("ParseFinishingStatusReferences: %v", err)
	}
	if len(statuses) != 5 {
		t.Fatalf("got %d statuses, want 5", len(statuses))
	}
	want := FinishingStatusReference{StatusId: "1", Status: "Finished", Count: 287}
	if statuses[0] != want {
		t.Errorf("first status = %+v, want %+v", statuses[0], want)
	}
}

func TestParseDriverReferences_RejectsMalformed(t *testing.T) {
	if _, err := ParseDriverReferences([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
