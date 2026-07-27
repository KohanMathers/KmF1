package feeds

import (
	"testing"

	"github.com/KohanMathers/KmF1/go/internal/testfixture"
)

func TestParseArchiveStatus_RealCapture(t *testing.T) {
	data := testfixture.Read(t, "testdata/archivestatus.json")

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
