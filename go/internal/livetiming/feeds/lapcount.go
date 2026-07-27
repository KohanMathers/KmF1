package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type SessionProgress struct {
	CurrentLap int
	TotalLaps  int
	Timestamp  time.Duration
}

type rawLapCount struct {
	CurrentLap int  `json:"CurrentLap"`
	TotalLaps  *int `json:"TotalLaps"`
}

func ParseLapCount(data []byte) ([]SessionProgress, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: LapCount: %w", err)
	}

	var d SessionProgressDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: LapCount: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type SessionProgressDecoder struct {
	progress      []SessionProgress
	lastTotalLaps int
}

func (d *SessionProgressDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw rawLapCount
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if raw.TotalLaps != nil {
		d.lastTotalLaps = *raw.TotalLaps
	}

	d.progress = append(d.progress, SessionProgress{
		CurrentLap: raw.CurrentLap,
		TotalLaps:  d.lastTotalLaps,
		Timestamp:  elapsed,
	})
	return nil
}

func (d *SessionProgressDecoder) Snapshot() []SessionProgress {
	out := make([]SessionProgress, len(d.progress))
	copy(out, d.progress)
	return out
}
