package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type RaceControlMessage struct {
	Message      string
	Category     string
	Flag         string
	Scope        string
	Sector       *int
	DriverNumber string
	Lap          int
	Status       string
	Timestamp    time.Duration
}

type rawRaceControlMessage struct {
	Category     string `json:"Category"`
	Flag         string `json:"Flag"`
	Scope        string `json:"Scope"`
	Sector       *int   `json:"Sector"`
	RacingNumber string `json:"RacingNumber"`
	Lap          int    `json:"Lap"`
	Status       string `json:"Status"`
	Message      string `json:"Message"`
}

func ParseRaceControlMessages(data []byte) ([]RaceControlMessage, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: RaceControlMessages: %w", err)
	}

	var d RaceControlMessageDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: RaceControlMessages: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type RaceControlMessageDecoder struct {
	messages []RaceControlMessage
}

func (d *RaceControlMessageDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	entries, err := decodeIndexedField(decoded, "Messages")
	if err != nil {
		return err
	}

	for _, idx := range sortedIndices(entries) {
		raw, err := json.Marshal(entries[idx])
		if err != nil {
			continue
		}
		var m rawRaceControlMessage
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		d.messages = append(d.messages, RaceControlMessage{
			Message:      m.Message,
			Category:     m.Category,
			Flag:         m.Flag,
			Scope:        m.Scope,
			Sector:       m.Sector,
			DriverNumber: m.RacingNumber,
			Lap:          m.Lap,
			Status:       m.Status,
			Timestamp:    elapsed,
		})
	}
	return nil
}

func (d *RaceControlMessageDecoder) Snapshot() []RaceControlMessage {
	out := make([]RaceControlMessage, len(d.messages))
	copy(out, d.messages)
	return out
}
