package wire

import "testing"

func TestMergeInto_PlainObjectMerge(t *testing.T) {
	dst := map[string]any{
		"1": map[string]any{"Name": "VER", "Points": float64(194)},
	}
	patch := map[string]any{
		"1": map[string]any{"Points": float64(212)},
		"4": map[string]any{"Name": "NOR"},
	}

	MergeInto(dst, patch)

	drv1 := dst["1"].(map[string]any)
	if drv1["Name"] != "VER" {
		t.Errorf(`dst["1"]["Name"] = %v, want "VER" (untouched field must survive the merge)`, drv1["Name"])
	}
	if drv1["Points"] != float64(212) {
		t.Errorf(`dst["1"]["Points"] = %v, want 212 (patched field must overwrite)`, drv1["Points"])
	}

	drv4, ok := dst["4"].(map[string]any)
	if !ok || drv4["Name"] != "NOR" {
		t.Errorf(`dst["4"] = %v, want a new entry with Name "NOR"`, dst["4"])
	}
}

func TestMergeInto_RealTeamRadioGrowsArray(t *testing.T) {
	data := readTestdata(t, "teamradio.jsonstream")
	records, err := SplitFrames(data)
	if err != nil {
		t.Fatalf("SplitFrames: %v", err)
	}
	if len(records) != 6 {
		t.Fatalf("got %d records, want 6", len(records))
	}

	state, err := Reconstruct(records, false)
	if err != nil {
		t.Fatalf("Reconstruct: %v", err)
	}

	captures, ok := state["Captures"].([]any)
	if !ok {
		t.Fatalf("Captures is %T, want []any", state["Captures"])
	}

	if len(captures) != 8 {
		t.Fatalf("got %d captures, want 8", len(captures))
	}

	wantRacingNumbers := []string{"27", "20", "63", "63", "44", "31", "16", "16"}
	for i, want := range wantRacingNumbers {
		clip, ok := captures[i].(map[string]any)
		if !ok {
			t.Fatalf("capture %d is %T, want map[string]any", i, captures[i])
		}
		if got := clip["RacingNumber"]; got != want {
			t.Errorf("capture %d: RacingNumber = %v, want %q", i, got, want)
		}
		if _, ok := clip["Path"].(string); !ok {
			t.Errorf("capture %d: missing Path field", i)
		}
	}
}

func TestMergeInto_RealTopThreeUpdatesFixedSlotsInPlace(t *testing.T) {
	data := readTestdata(t, "topthree.jsonstream")
	records, err := SplitFrames(data)
	if err != nil {
		t.Fatalf("SplitFrames: %v", err)
	}
	if len(records) != 6 {
		t.Fatalf("got %d records, want 6", len(records))
	}

	state, err := Reconstruct(records, false)
	if err != nil {
		t.Fatalf("Reconstruct: %v", err)
	}

	lines, ok := state["Lines"].([]any)
	if !ok {
		t.Fatalf("Lines is %T, want []any", state["Lines"])
	}

	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3 (slots must update in place, not grow)", len(lines))
	}

	type want struct {
		racingNumber string
		tla          string
		lapState     float64
		diffToAhead  string
	}
	wants := []want{
		{"4", "NOR", 64, "LAP 1"},
		{"1", "VER", 80, ""},
		{"44", "HAM", 64, ""},
	}

	for i, w := range wants {
		line, ok := lines[i].(map[string]any)
		if !ok {
			t.Fatalf("line %d is %T, want map[string]any", i, lines[i])
		}
		if got := line["RacingNumber"]; got != w.racingNumber {
			t.Errorf("line %d: RacingNumber = %v, want %q", i, got, w.racingNumber)
		}
		if got := line["Tla"]; got != w.tla {
			t.Errorf("line %d: Tla = %v, want %q (must survive from the initial snapshot)", i, got, w.tla)
		}
		if got := line["LapState"]; got != w.lapState {
			t.Errorf("line %d: LapState = %v, want %v", i, got, w.lapState)
		}
		if w.diffToAhead != "" {
			if got := line["DiffToAhead"]; got != w.diffToAhead {
				t.Errorf("line %d: DiffToAhead = %v, want %q", i, got, w.diffToAhead)
			}
		}
	}
}

func TestMergeInto_ArrayPatchFillsGap(t *testing.T) {
	dst := map[string]any{
		"Items": []any{"a"},
	}

	patch := map[string]any{
		"Items": map[string]any{"2": "c"},
	}

	MergeInto(dst, patch)

	items := dst["Items"].([]any)
	if len(items) != 3 {
		t.Fatalf("got %d items, want 3", len(items))
	}
	if items[0] != "a" || items[1] != nil || items[2] != "c" {
		t.Fatalf("got %v", items)
	}
}

func TestReconstruct_RejectsNonObjectTopLevel(t *testing.T) {
	records := []Record{{Payload: `["not", "an", "object"]`}}
	if _, err := Reconstruct(records, false); err == nil {
		t.Fatal("expected an error when a record's top-level payload isn't a JSON object")
	}
}
