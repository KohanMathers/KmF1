package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type SessionClock struct {
	Remaining     time.Duration
	Extrapolating bool
	Timestamp     time.Duration
}

type rawSessionClock struct {
	Remaining     string `json:"Remaining"`
	Extrapolating *bool  `json:"Extrapolating"`
}

func ParseSessionClock(data []byte) ([]SessionClock, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: ExtrapolatedClock: %w", err)
	}

	var d SessionClockDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: ExtrapolatedClock: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type SessionClockDecoder struct {
	clocks            []SessionClock
	lastExtrapolating bool
}

func (d *SessionClockDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw rawSessionClock
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if raw.Extrapolating != nil {
		d.lastExtrapolating = *raw.Extrapolating
	}

	remaining, _ := parseF1Duration(raw.Remaining)
	d.clocks = append(d.clocks, SessionClock{
		Remaining:     remaining,
		Extrapolating: d.lastExtrapolating,
		Timestamp:     elapsed,
	})
	return nil
}

func (d *SessionClockDecoder) Snapshot() []SessionClock {
	out := make([]SessionClock, len(d.clocks))
	copy(out, d.clocks)
	return out
}
