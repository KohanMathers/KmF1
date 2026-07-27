package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type TrackStatus struct {
	Status    string
	Message   string
	Timestamp time.Duration
}

type rawTrackStatus struct {
	Status  string `json:"Status"`
	Message string `json:"Message"`
}

func ParseTrackStatus(data []byte) ([]TrackStatus, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TrackStatus: %w", err)
	}

	var d TrackStatusDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TrackStatus: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type TrackStatusDecoder struct {
	events []TrackStatus
}

func (d *TrackStatusDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw rawTrackStatus
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	d.events = append(d.events, TrackStatus{
		Status:    raw.Status,
		Message:   raw.Message,
		Timestamp: elapsed,
	})
	return nil
}

func (d *TrackStatusDecoder) Snapshot() []TrackStatus {
	out := make([]TrackStatus, len(d.events))
	copy(out, d.events)
	return out
}
