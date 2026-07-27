package feeds

import (
	"os"
	"testing"
	"time"
)

func TestParseSessionInfo_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/sessioninfo.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	info, err := ParseSessionInfo(data)
	if err != nil {
		t.Fatalf("ParseSessionInfo: %v", err)
	}

	if info.Key != 9539 {
		t.Errorf("Key = %d, want 9539", info.Key)
	}
	if info.Type != "Race" {
		t.Errorf("Type = %q, want %q", info.Type, "Race")
	}
	if info.Name != "Race" {
		t.Errorf("Name = %q, want %q", info.Name, "Race")
	}
	if info.Path != "2024/2024-06-23_Spanish_Grand_Prix/2024-06-23_Race/" {
		t.Errorf("Path = %q", info.Path)
	}

	wantStart := time.Date(2024, 6, 23, 15, 0, 0, 0, time.UTC)
	if !info.StartDate.Equal(wantStart) {
		t.Errorf("StartDate = %v, want %v", info.StartDate, wantStart)
	}
	wantEnd := time.Date(2024, 6, 23, 17, 0, 0, 0, time.UTC)
	if !info.EndDate.Equal(wantEnd) {
		t.Errorf("EndDate = %v, want %v", info.EndDate, wantEnd)
	}
	if info.GmtOffset != 2*time.Hour {
		t.Errorf("GmtOffset = %v, want 2h", info.GmtOffset)
	}

	if info.Meeting.Key != 1238 {
		t.Errorf("Meeting.Key = %d, want 1238", info.Meeting.Key)
	}
	if info.Meeting.Name != "Spanish Grand Prix" {
		t.Errorf("Meeting.Name = %q", info.Meeting.Name)
	}
	if info.Meeting.OfficialName != "FORMULA 1 ARAMCO GRAN PREMIO DE ESPAÑA 2024" {
		t.Errorf("Meeting.OfficialName = %q", info.Meeting.OfficialName)
	}
	if info.Meeting.Location != "Barcelona" {
		t.Errorf("Meeting.Location = %q", info.Meeting.Location)
	}
	if info.Meeting.Country.Code != "ESP" {
		t.Errorf("Meeting.Country.Code = %q, want %q", info.Meeting.Country.Code, "ESP")
	}
	if info.Meeting.Country.Name != "Spain" {
		t.Errorf("Meeting.Country.Name = %q", info.Meeting.Country.Name)
	}
	if info.Meeting.Circuit.ShortName != "Catalunya" {
		t.Errorf("Meeting.Circuit.ShortName = %q, want %q", info.Meeting.Circuit.ShortName, "Catalunya")
	}

	if info.ArchiveStatus.Status != "Generating" {
		t.Errorf("ArchiveStatus.Status = %q, want %q (frozen at session start)",
			info.ArchiveStatus.Status, "Generating")
	}
}

func TestParseSessionInfo_RejectsEmptyData(t *testing.T) {
	if _, err := ParseSessionInfo([]byte{}); err == nil {
		t.Fatal("expected an error for empty input")
	}
}

func TestParseSessionInfo_RejectsMalformedJSON(t *testing.T) {
	_, err := ParseSessionInfo([]byte("00:00:00.000not json\r\n"))
	if err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
