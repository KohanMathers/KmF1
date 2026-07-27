package feeds

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type AudioStream struct {
	Name     string
	Language string
	Uri      string
	Path     string
	Utc      time.Time
}

type ContentStream struct {
	Type     string
	Name     string
	Language string
	Uri      string
	Path     string
	Utc      time.Time
}

type rawAudioStream struct {
	Name     string `json:"Name"`
	Language string `json:"Language"`
	Uri      string `json:"Uri"`
	Path     string `json:"Path"`
	Utc      string `json:"Utc"`
}

type rawContentStream struct {
	Type     string `json:"Type"`
	Name     string `json:"Name"`
	Language string `json:"Language"`
	Uri      string `json:"Uri"`
	Path     string `json:"Path"`
	Utc      string `json:"Utc"`
}

func ParseAudioStreams(data []byte) ([]AudioStream, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: AudioStreams: %w", err)
	}

	var d AudioStreamDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: AudioStreams: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type AudioStreamDecoder struct {
	streams []AudioStream
}

func (d *AudioStreamDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	entries, err := decodeIndexedField(decoded, "Streams")
	if err != nil {
		return err
	}

	for _, idx := range sortedIndices(entries) {
		raw, err := json.Marshal(entries[idx])
		if err != nil {
			continue
		}
		var s rawAudioStream
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		utc, _ := time.Parse(time.RFC3339Nano, s.Utc)
		d.streams = append(d.streams, AudioStream{
			Name: s.Name, Language: s.Language, Uri: s.Uri, Path: s.Path, Utc: utc,
		})
	}
	return nil
}

func (d *AudioStreamDecoder) Snapshot() []AudioStream {
	out := make([]AudioStream, len(d.streams))
	copy(out, d.streams)
	return out
}

func ParseContentStreams(data []byte) ([]ContentStream, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: ContentStreams: %w", err)
	}

	var d ContentStreamDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: ContentStreams: record %d: %w", i, err)
		}
	}
	return d.Snapshot(), nil
}

type ContentStreamDecoder struct {
	streams []ContentStream
}

func (d *ContentStreamDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	entries, err := decodeIndexedField(decoded, "Streams")
	if err != nil {
		return err
	}

	for _, idx := range sortedIndices(entries) {
		raw, err := json.Marshal(entries[idx])
		if err != nil {
			continue
		}
		var s rawContentStream
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		utc, _ := time.Parse(time.RFC3339Nano, s.Utc)
		d.streams = append(d.streams, ContentStream{
			Type: s.Type, Name: s.Name, Language: s.Language, Uri: s.Uri, Path: s.Path, Utc: utc,
		})
	}
	return nil
}

func (d *ContentStreamDecoder) Snapshot() []ContentStream {
	out := make([]ContentStream, len(d.streams))
	copy(out, d.streams)
	return out
}

func decodeIndexedField(decoded []byte, fieldName string) (map[int]map[string]any, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(decoded, &top); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	var container any
	if raw, ok := top[fieldName]; ok {
		if err := json.Unmarshal(raw, &container); err != nil {
			return nil, fmt.Errorf("%s: %w", fieldName, err)
		}
	}

	return indexedEntries(container), nil
}

func sortedIndices(entries map[int]map[string]any) []int {
	indices := make([]int, 0, len(entries))
	for idx := range entries {
		indices = append(indices, idx)
	}
	sort.Ints(indices)
	return indices
}
