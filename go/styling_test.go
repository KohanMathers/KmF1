package kmf1

import "testing"

func TestSeason_TeamStyle_RealLiveFeedName(t *testing.T) {
	s := &Season{year: 2024}

	got, ok := s.TeamStyle("Red Bull Racing")
	if !ok {
		t.Fatal(`TeamStyle("Red Bull Racing") not found`)
	}
	if got.ShortName != "Red Bull" {
		t.Errorf("ShortName = %q, want %q", got.ShortName, "Red Bull")
	}
	if got.Colors.Official != "#3671c6" {
		t.Errorf("Colors.Official = %q, want %q", got.Colors.Official, "#3671c6")
	}
}

func TestSeason_TeamStyle_UnknownYear(t *testing.T) {
	s := &Season{year: 1994}
	if _, ok := s.TeamStyle("Williams"); ok {
		t.Error("TeamStyle found for a year with no curated data, want false")
	}
}

func TestSeason_TeamStyles_NonEmpty(t *testing.T) {
	s := &Season{year: 2024}
	teams := s.TeamStyles()
	if len(teams) != 10 {
		t.Fatalf("TeamStyles() = %d entries, want 10", len(teams))
	}
}

func TestSeason_CompoundColor(t *testing.T) {
	s := &Season{year: 2024}
	got, ok := s.CompoundColor("soft")
	if !ok || got != "#da291c" {
		t.Errorf(`CompoundColor("soft") = (%q, %v), want ("#da291c", true)`, got, ok)
	}
}

func TestSeason_Compounds_NonEmpty(t *testing.T) {
	s := &Season{year: 2024}
	if len(s.Compounds()) == 0 {
		t.Error("expected non-empty Compounds()")
	}
}
