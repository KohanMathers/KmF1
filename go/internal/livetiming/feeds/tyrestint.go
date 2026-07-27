package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type TyreStint struct {
	DriverNumber string
	Compound     string
	New          bool
	TyresChanged bool
	Stint        int
	StintAge     int
	StartLap     int
	LapFlags     int
	Timestamp    time.Duration
}

type stintBuilder struct {
	compound     string
	new          bool
	tyresChanged bool
	stintAge     int
	startLap     int
	lapFlags     int
}

func ParseTyreStints(data []byte) (map[string][]TyreStint, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TyreStint: %w", err)
	}

	var d TyreStintDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TyreStint: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type TyreStintDecoder struct {
	builders map[string]map[int]*stintBuilder
	result   map[string][]TyreStint
}

func (d *TyreStintDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var top struct {
		Lines map[string]map[string]any `json:"Lines"`
	}
	if err := json.Unmarshal(decoded, &top); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if d.builders == nil {
		d.builders = make(map[string]map[int]*stintBuilder)
		d.result = make(map[string][]TyreStint)
	}

	for drv, patch := range top.Lines {
		stints := indexedEntries(patch["Stints"])
		if len(stints) == 0 {
			continue
		}

		if _, ok := d.builders[drv]; !ok {
			d.builders[drv] = make(map[int]*stintBuilder)
		}

		for idx, fields := range stints {
			b, ok := d.builders[drv][idx]
			if !ok {
				b = &stintBuilder{}
				d.builders[drv][idx] = b
			}
			applyStintFields(fields, b)

			d.result[drv] = append(d.result[drv], TyreStint{
				DriverNumber: drv,
				Compound:     b.compound,
				New:          b.new,
				TyresChanged: b.tyresChanged,
				Stint:        idx + 1,
				StintAge:     b.stintAge,
				StartLap:     b.startLap,
				LapFlags:     b.lapFlags,
				Timestamp:    elapsed,
			})
		}
	}
	return nil
}

func (d *TyreStintDecoder) Snapshot() map[string][]TyreStint {
	out := make(map[string][]TyreStint, len(d.result))
	for drv, snapshots := range d.result {
		cp := make([]TyreStint, len(snapshots))
		copy(cp, snapshots)
		out[drv] = cp
	}
	return out
}

func applyStintFields(fields map[string]any, b *stintBuilder) {
	if v, ok := fields["Compound"].(string); ok {
		b.compound = v
	}
	if v, ok := fields["New"].(string); ok {
		b.new = v == "true"
	}
	if v, ok := fields["TyresNotChanged"].(string); ok {
		b.tyresChanged = v == "0"
	}
	if v, ok := fields["TotalLaps"]; ok {
		b.stintAge = int(toFloat(v))
	}
	if v, ok := fields["StartLaps"]; ok {
		b.startLap = int(toFloat(v))
	}
	if v, ok := fields["LapFlags"]; ok {
		b.lapFlags = int(toFloat(v))
	}
}
