package feeds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type TopThreeLive struct {
	Position        int
	DriverNumber    string
	Abbreviation    string
	BroadcastName   string
	FullName        string
	TeamName        string
	TeamColour      string
	LapTime         *time.Duration
	LapState        int
	DiffToAhead     string
	DiffToLeader    string
	OverallFastest  bool
	PersonalFastest bool
	Timestamp       time.Duration
}

type topThreeSlotBuilder struct {
	TopThreeLive
}

func ParseTopThreeLive(data []byte) (map[int][]TopThreeLive, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TopThree: %w", err)
	}

	var d TopThreeLiveDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TopThree: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type TopThreeLiveDecoder struct {
	builders map[int]*topThreeSlotBuilder
	result   map[int][]TopThreeLive
}

func (d *TopThreeLiveDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var top struct {
		Lines json.RawMessage `json:"Lines"`
	}
	if err := json.Unmarshal(decoded, &top); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	var container any
	if err := json.Unmarshal(top.Lines, &container); err != nil {
		return fmt.Errorf("lines: %w", err)
	}

	if d.builders == nil {
		d.builders = make(map[int]*topThreeSlotBuilder)
		d.result = make(map[int][]TopThreeLive)
	}
	for slot, fields := range indexedEntries(container) {
		b, ok := d.builders[slot]
		if !ok {
			b = &topThreeSlotBuilder{}
			d.builders[slot] = b
		}
		if applyTopThreeFields(fields, &b.TopThreeLive) {
			snapshot := b.TopThreeLive
			snapshot.Timestamp = elapsed
			d.result[slot] = append(d.result[slot], snapshot)
		}
	}
	return nil
}

func (d *TopThreeLiveDecoder) Snapshot() map[int][]TopThreeLive {
	out := make(map[int][]TopThreeLive, len(d.result))
	for slot, snapshots := range d.result {
		cp := make([]TopThreeLive, len(snapshots))
		copy(cp, snapshots)
		out[slot] = cp
	}
	return out
}

func applyTopThreeFields(fields map[string]any, t *TopThreeLive) bool {
	changed := false

	if v, ok := fields["Position"].(string); ok {
		if n, err := strconv.Atoi(v); err == nil {
			t.Position = n
			changed = true
		}
	}
	if v, ok := fields["RacingNumber"].(string); ok {
		t.DriverNumber = v
		changed = true
	}
	if v, ok := fields["Tla"].(string); ok {
		t.Abbreviation = v
		changed = true
	}
	if v, ok := fields["BroadcastName"].(string); ok {
		t.BroadcastName = v
		changed = true
	}
	if v, ok := fields["FullName"].(string); ok {
		t.FullName = v
		changed = true
	}
	if v, ok := fields["Team"].(string); ok {
		t.TeamName = v
		changed = true
	}
	if v, ok := fields["TeamColour"].(string); ok {
		t.TeamColour = v
		changed = true
	}
	if v, ok := fields["LapTime"].(string); ok {
		changed = true
		if v == "" {
			t.LapTime = nil
		} else if dur, err := parseF1Duration(v); err == nil {
			t.LapTime = &dur
		}
	}
	if v, ok := fields["LapState"]; ok {
		t.LapState = int(toFloat(v))
		changed = true
	}
	if v, ok := fields["DiffToAhead"].(string); ok {
		t.DiffToAhead = v
		changed = true
	}
	if v, ok := fields["DiffToLeader"].(string); ok {
		t.DiffToLeader = v
		changed = true
	}
	if v, ok := fields["OverallFastest"].(bool); ok {
		t.OverallFastest = v
		changed = true
	}
	if v, ok := fields["PersonalFastest"].(bool); ok {
		t.PersonalFastest = v
		changed = true
	}

	return changed
}
