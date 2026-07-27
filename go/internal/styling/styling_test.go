package styling

import "testing"

func TestForYear_KnownYear(t *testing.T) {
	s, ok := ForYear(2024)
	if !ok {
		t.Fatal("ForYear(2024) not found, want the vendored 2024 entry")
	}
	if len(s.teams) == 0 {
		t.Error("expected a non-empty team table for 2024")
	}
	if len(s.compoundColors) == 0 {
		t.Error("expected a non-empty compound color table for 2024")
	}
}

func TestForYear_UnknownYear(t *testing.T) {
	if _, ok := ForYear(1994); ok {
		t.Error("ForYear(1994) found, want false — no curated data for this year")
	}
}

func TestSeasonConstants_CompoundColor(t *testing.T) {
	s, _ := ForYear(2024)

	tests := []struct {
		compound string
		want     string
	}{
		{"SOFT", "#da291c"},
		{"soft", "#da291c"},
		{"Medium", "#ffd12e"},
		{"HARD", "#f0f0ec"},
	}
	for _, tt := range tests {
		got, ok := s.CompoundColor(tt.compound)
		if !ok || got != tt.want {
			t.Errorf("CompoundColor(%q) = (%q, %v), want (%q, true)", tt.compound, got, ok, tt.want)
		}
	}

	if _, ok := s.CompoundColor("HYPERSOFT"); ok {
		t.Error(`CompoundColor("HYPERSOFT") found in 2024 data, want false (retired compound)`)
	}
}

func TestSeasonConstants_Compounds_Sorted(t *testing.T) {
	s, _ := ForYear(2024)
	compounds := s.Compounds()
	if len(compounds) == 0 {
		t.Fatal("expected non-empty Compounds()")
	}
	for i := 1; i < len(compounds); i++ {
		if compounds[i-1] > compounds[i] {
			t.Errorf("Compounds() not sorted: %v", compounds)
			break
		}
	}
}

func TestSeasonConstants_Team_RealLiveFeedNames(t *testing.T) {
	s, ok := ForYear(2024)
	if !ok {
		t.Fatal("ForYear(2024) not found")
	}

	tests := []struct {
		liveTeamName  string
		wantShortName string
	}{
		{"Red Bull Racing", "Red Bull"},
		{"Haas F1 Team", "Haas"},
		{"Kick Sauber", "Sauber"},
		{"RB", "RB"},
		{"Aston Martin", "Aston Martin"},
		{"Ferrari", "Ferrari"},
		{"McLaren", "McLaren"},
		{"Mercedes", "Mercedes"},
		{"Williams", "Williams"},
		{"Alpine", "Alpine"},
	}
	for _, tt := range tests {
		got, ok := s.Team(tt.liveTeamName)
		if !ok {
			t.Errorf("Team(%q) not found, want short name %q", tt.liveTeamName, tt.wantShortName)
			continue
		}
		if got.ShortName != tt.wantShortName {
			t.Errorf("Team(%q).ShortName = %q, want %q", tt.liveTeamName, got.ShortName, tt.wantShortName)
		}
	}
}

func TestSeasonConstants_Team_CuratedKeyAndCaseInsensitive(t *testing.T) {
	s, _ := ForYear(2024)

	got, ok := s.Team("RED BULL")
	if !ok || got.ShortName != "Red Bull" {
		t.Errorf(`Team("RED BULL") = (%+v, %v), want Red Bull`, got, ok)
	}

	got, ok = s.Team("red bull")
	if !ok || got.Colors.Official != "#3671c6" || got.Colors.FastF1 != "#0600ef" {
		t.Errorf(`Team("red bull").Colors = %+v, want official #3671c6 / fastf1 #0600ef`, got.Colors)
	}
}

func TestSeasonConstants_Team_NotFound(t *testing.T) {
	s, _ := ForYear(2024)
	if _, ok := s.Team("Definitely Not A Real Team"); ok {
		t.Error("Team(...) found for a nonsense identifier, want false")
	}
	if _, ok := s.Team(""); ok {
		t.Error(`Team("") found, want false`)
	}
}

func TestSeasonConstants_Teams_SortedByShortName(t *testing.T) {
	s, _ := ForYear(2024)
	teams := s.Teams()
	if len(teams) != 10 {
		t.Fatalf("Teams() = %d entries, want 10 (2024 grid)", len(teams))
	}
	for i := 1; i < len(teams); i++ {
		if teams[i-1].ShortName > teams[i].ShortName {
			t.Errorf("Teams() not sorted by ShortName: %+v", teams)
			break
		}
	}
}

func TestSeasonConstants_HistoricalTeamNames(t *testing.T) {
	s2018, ok := ForYear(2018)
	if !ok {
		t.Fatal("ForYear(2018) not found")
	}
	if _, ok := s2018.Team("Force India"); !ok {
		t.Error(`2018: Team("Force India") not found`)
	}
	if _, ok := s2018.Team("Toro Rosso"); !ok {
		t.Error(`2018: Team("Toro Rosso") not found`)
	}

	s2026, ok := ForYear(2026)
	if !ok {
		t.Fatal("ForYear(2026) not found")
	}
	if _, ok := s2026.Team("Audi"); !ok {
		t.Error(`2026: Team("Audi") not found`)
	}
	if _, ok := s2026.Team("Cadillac"); !ok {
		t.Error(`2026: Team("Cadillac") not found`)
	}
}
