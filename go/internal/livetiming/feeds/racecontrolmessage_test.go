package feeds

import (
	"testing"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseRaceControlMessages_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/racecontrolmessages.jsonstream")

	messages, err := ParseRaceControlMessages(data)
	if err != nil {
		t.Fatalf("ParseRaceControlMessages: %v", err)
	}
	if len(messages) != 105 {
		t.Fatalf("got %d messages, want 105", len(messages))
	}

	first := messages[0]
	if first.Message != "PINK HEAD PADDING MATERIAL MUST BE USED" {
		t.Errorf("first message = %q, want the padding-material notice", first.Message)
	}
	if first.Category != "Other" || first.Lap != 1 {
		t.Errorf("first message Category/Lap = %q/%d, want Other/1", first.Category, first.Lap)
	}
	if first.Timestamp != time.Minute+20*time.Second+891*time.Millisecond {
		t.Errorf("first message Timestamp = %v, want 00:01:20.891", first.Timestamp)
	}

	var foundDeletion bool
	for _, m := range messages {
		if m.Message == "CAR 55 (SAI) TIME 1:20.046 DELETED - TRACK LIMITS AT TURN 2 LAP 3 15:06:13" {
			foundDeletion = true
			if m.Category != "Other" {
				t.Errorf("deletion message Category = %q, want Other (deletions are NOT a dedicated category)", m.Category)
			}
			if m.DriverNumber != "" {
				t.Errorf("deletion message DriverNumber = %q, want empty (driver number is embedded in Message text, not a structured field here)", m.DriverNumber)
			}
		}
	}
	if !foundDeletion {
		t.Fatal("expected to find the real car 55 deletion message")
	}

	var foundBlueFlag bool
	for _, m := range messages {
		if m.Flag == "BLUE" && m.DriverNumber == "22" {
			foundBlueFlag = true
			if m.Scope != "Driver" {
				t.Errorf("blue flag Scope = %q, want Driver", m.Scope)
			}
			if m.Category != "Flag" {
				t.Errorf("blue flag Category = %q, want Flag", m.Category)
			}
		}
	}
	if !foundBlueFlag {
		t.Fatal("expected to find a real blue-flag message for driver 22")
	}

	var foundDRS bool
	for _, m := range messages {
		if m.Category == "Drs" && m.Status == "DISABLED" {
			foundDRS = true
		}
	}
	if !foundDRS {
		t.Fatal("expected to find a real DRS-disabled message")
	}
}

func TestParseRaceControlMessages_RejectsMalformed(t *testing.T) {
	if _, err := ParseRaceControlMessages([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
