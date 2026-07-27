package feeds

import (
	"os"
	"testing"
)

func TestParseArchiveStatus_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/archivestatus.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	status, err := ParseArchiveStatus(data)
	if err != nil {
		t.Fatalf("ParseArchiveStatus: %v", err)
	}
	if status.Status != "Complete" {
		t.Errorf("Status = %q, want %q", status.Status, "Complete")
	}
}

func TestParseArchiveStatus_RejectsMalformed(t *testing.T) {
	if _, err := ParseArchiveStatus([]byte("not json")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
