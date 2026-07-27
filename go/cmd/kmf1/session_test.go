package main

import "testing"

func TestResolveSessionType_CaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"race", "Race"},
		{"RACE", "Race"},
		{"Race", "Race"},
		{"qualifying", "Qualifying"},
		{"sprint qualifying", "Sprint Qualifying"},
		{"SPRINT QUALIFYING", "Sprint Qualifying"},
		{"sprint", "Sprint"},
		{"practice 1", "Practice 1"},
		{"PRACTICE 3", "Practice 3"},
	}
	for _, tt := range tests {
		got, err := resolveSessionType(tt.input)
		if err != nil {
			t.Errorf("resolveSessionType(%q) error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("resolveSessionType(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveSessionType_Unknown(t *testing.T) {
	if _, err := resolveSessionType("Endurance"); err == nil {
		t.Error("resolveSessionType(\"Endurance\") returned no error, want one")
	}
}

func TestEventIdentifier_IntVsString(t *testing.T) {
	if got := eventIdentifier("10"); got != 10 {
		t.Errorf("eventIdentifier(\"10\") = %#v (%T), want int 10", got, got)
	}
	if got := eventIdentifier("Spanish Grand Prix"); got != "Spanish Grand Prix" {
		t.Errorf("eventIdentifier(\"Spanish Grand Prix\") = %#v, want the string unchanged", got)
	}

	if got := eventIdentifier("05"); got != 5 {
		t.Errorf("eventIdentifier(\"05\") = %#v, want int 5", got)
	}
}
