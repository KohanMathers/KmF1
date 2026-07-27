package live

import (
	"testing"
	"time"
)

func TestTimeAnchor_FirstTimestampBecomesZero(t *testing.T) {
	var a TimeAnchor
	if a.Known() {
		t.Fatal("Known() = true before any timestamp fed, want false")
	}

	got := a.Elapsed("2024-06-23T14:00:00Z")
	if got != 0 {
		t.Errorf("Elapsed() on the first real timestamp = %v, want 0", got)
	}
	if !a.Known() {
		t.Error("Known() = false after a real timestamp was fed, want true")
	}
}

func TestTimeAnchor_SubsequentTimestampsAreRelative(t *testing.T) {
	var a TimeAnchor
	a.Elapsed("2024-06-23T14:00:00Z")

	got := a.Elapsed("2024-06-23T14:00:05.500Z")
	want := 5500 * time.Millisecond
	if got != want {
		t.Errorf("Elapsed() 5.5s later = %v, want %v", got, want)
	}
}

func TestTimeAnchor_EmptyTimestampAlwaysZero(t *testing.T) {
	var a TimeAnchor
	if got := a.Elapsed(""); got != 0 {
		t.Errorf(`Elapsed("") before any anchor = %v, want 0`, got)
	}
	if a.Known() {
		t.Error("an empty timestamp must not establish the anchor")
	}

	a.Elapsed("2024-06-23T14:00:00Z")
	if got := a.Elapsed(""); got != 0 {
		t.Errorf(`Elapsed("") after the anchor is known = %v, want 0`, got)
	}
}

func TestTimeAnchor_MalformedTimestampDegradesToZero(t *testing.T) {
	var a TimeAnchor
	if got := a.Elapsed("not a timestamp"); got != 0 {
		t.Errorf("Elapsed(malformed) = %v, want 0", got)
	}
	if a.Known() {
		t.Error("a malformed timestamp must not establish the anchor")
	}
}

func TestTimeAnchor_TimestampBeforeAnchorIsNegative(t *testing.T) {

	var a TimeAnchor
	a.Elapsed("2024-06-23T14:00:10Z")

	got := a.Elapsed("2024-06-23T14:00:05Z")
	want := -5 * time.Second
	if got != want {
		t.Errorf("Elapsed() for an earlier timestamp = %v, want %v", got, want)
	}
}
