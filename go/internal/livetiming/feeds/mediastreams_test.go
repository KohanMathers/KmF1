package feeds

import (
	"os"
	"testing"
	"time"
)

func TestParseAudioStreams_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/audiostreams.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	streams, err := ParseAudioStreams(data)
	if err != nil {
		t.Fatalf("ParseAudioStreams: %v", err)
	}
	if len(streams) != 1 {
		t.Fatalf("got %d streams, want 1 (only English world feed observed)", len(streams))
	}

	s := streams[0]
	if s.Name != "Live coverage (EN)" || s.Language != "en" {
		t.Errorf("Name/Language = %q/%q", s.Name, s.Language)
	}
	if s.Uri != "https://dtksvu0irgynk.cloudfront.net/out/v1/7ed89990eb2a4d39870ff19793519937/index.m3u8" {
		t.Errorf("Uri = %q", s.Uri)
	}
	if s.Path != "Live_coverage_(EN)-en/stream.m3u8" {
		t.Errorf("Path = %q", s.Path)
	}
	wantUtc, _ := time.Parse(time.RFC3339Nano, "2024-06-23T12:06:17.317Z")
	if !s.Utc.Equal(wantUtc) {
		t.Errorf("Utc = %v, want %v", s.Utc, wantUtc)
	}
}

func TestParseContentStreams_RealCapture(t *testing.T) {
	data, err := os.ReadFile("testdata/contentstreams.jsonstream")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	streams, err := ParseContentStreams(data)
	if err != nil {
		t.Fatalf("ParseContentStreams: %v", err)
	}
	if len(streams) != 2 {
		t.Fatalf("got %d streams, want 2 (one Commentary, one Audio)", len(streams))
	}

	commentary := streams[0]
	if commentary.Type != "Commentary" || commentary.Name != "monterosa" {
		t.Errorf("commentary entry = %+v", commentary)
	}
	if commentary.Path != "" {
		t.Errorf("commentary Path = %q, want empty (confirmed real: absent entirely for this entry type)", commentary.Path)
	}
	if !commentary.Utc.IsZero() {
		t.Errorf("commentary Utc = %v, want zero value (confirmed real placeholder)", commentary.Utc)
	}

	audio := streams[1]
	if audio.Type != "Audio" || audio.Name != "Live coverage (EN)" {
		t.Errorf("audio entry = %+v", audio)
	}
	if audio.Utc.IsZero() {
		t.Error("audio entry Utc = zero, want a real timestamp")
	}
}

func TestParseAudioStreams_RejectsMalformed(t *testing.T) {
	if _, err := ParseAudioStreams([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}

func TestParseContentStreams_RejectsMalformed(t *testing.T) {
	if _, err := ParseContentStreams([]byte("00:00:00.000not json\r\n")); err == nil {
		t.Fatal("expected an error for a malformed payload")
	}
}
