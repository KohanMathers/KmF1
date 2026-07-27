package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type SessionStatusEvent struct {
	Status    string
	Timestamp time.Duration
}

type rawSessionStatusEvent struct {
	Status string `json:"Status"`
}

func ParseSessionStatus(data []byte) ([]SessionStatusEvent, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: SessionStatus: %w", err)
	}

	var d SessionStatusDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: SessionStatus: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type SessionStatusDecoder struct {
	events []SessionStatusEvent
}

func (d *SessionStatusDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw rawSessionStatusEvent
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	d.events = append(d.events, SessionStatusEvent{
		Status:    raw.Status,
		Timestamp: elapsed,
	})
	return nil
}

func (d *SessionStatusDecoder) Snapshot() []SessionStatusEvent {
	out := make([]SessionStatusEvent, len(d.events))
	copy(out, d.events)
	return out
}
