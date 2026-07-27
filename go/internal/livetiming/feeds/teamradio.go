package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type TeamRadio struct {
	DriverNumber string
	Path         string
	Utc          time.Time

	Timestamp time.Duration
}

type rawTeamRadioCapture struct {
	Utc          string `json:"Utc"`
	RacingNumber string `json:"RacingNumber"`
	Path         string `json:"Path"`
}

func ParseTeamRadio(data []byte) ([]TeamRadio, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TeamRadio: %w", err)
	}

	var d TeamRadioDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TeamRadio: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type TeamRadioDecoder struct {
	captures []TeamRadio
}

func (d *TeamRadioDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	entries, err := decodeIndexedField(decoded, "Captures")
	if err != nil {
		return err
	}

	for _, idx := range sortedIndices(entries) {
		raw, err := json.Marshal(entries[idx])
		if err != nil {
			continue
		}
		var c rawTeamRadioCapture
		if err := json.Unmarshal(raw, &c); err != nil {
			continue
		}
		utc, err := time.Parse(time.RFC3339Nano, c.Utc)
		if err != nil {
			continue
		}
		d.captures = append(d.captures, TeamRadio{
			DriverNumber: c.RacingNumber,
			Path:         c.Path,
			Utc:          utc,
			Timestamp:    elapsed,
		})
	}
	return nil
}

func (d *TeamRadioDecoder) Snapshot() []TeamRadio {
	out := make([]TeamRadio, len(d.captures))
	copy(out, d.captures)
	return out
}
