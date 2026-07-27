package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type TimingPatch struct {
	Elapsed time.Duration
	Data    map[string]any
}

func SplitTimingDataByDriver(data []byte) (map[string][]TimingPatch, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: TimingData: %w", err)
	}

	var d TimingDataDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: TimingData: record %d: %w", i, err)
		}
	}
	return d.Patches(), nil
}

type TimingDataDecoder struct {
	perDriver map[string][]TimingPatch
}

func (d *TimingDataDecoder) Feed(elapsed time.Duration, payload string) error {
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

	if d.perDriver == nil {
		d.perDriver = make(map[string][]TimingPatch)
	}
	for drv, patch := range top.Lines {
		d.perDriver[drv] = append(d.perDriver[drv], TimingPatch{
			Elapsed: elapsed,
			Data:    patch,
		})
	}
	return nil
}

func (d *TimingDataDecoder) Patches() map[string][]TimingPatch {
	out := make(map[string][]TimingPatch, len(d.perDriver))
	for drv, patches := range d.perDriver {
		cp := make([]TimingPatch, len(patches))
		copy(cp, patches)
		out[drv] = cp
	}
	return out
}

func (d *TimingDataDecoder) Laps() map[string][]LapData {
	result := make(map[string][]LapData, len(d.perDriver))
	for drv, patches := range d.perDriver {
		laps, personalBests := buildRawLaps(drv, patches)
		laps, _ = postProcess(laps, personalBests)

		lapData := make([]LapData, 0, len(laps))
		for _, l := range laps {
			if l == nil {
				continue
			}
			lapData = append(lapData, l.finalize())
		}
		result[drv] = lapData
	}
	return result
}

func (d *TimingDataDecoder) Intervals() map[string][]TimingInterval {
	result := make(map[string][]TimingInterval, len(d.perDriver))
	for drv, patches := range d.perDriver {
		result[drv] = timingIntervalsForDriver(drv, patches)
	}
	return result
}
