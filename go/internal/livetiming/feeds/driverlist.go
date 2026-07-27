package feeds

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type DriverInfo struct {
	DriverNumber  string
	Abbreviation  string
	BroadcastName string
	FullName      string
	TeamName      string
	TeamColour    string
	Reference     string
	HeadshotUrl   string
	CountryCode   string
	Line          int
}

type rawDriverInfo struct {
	RacingNumber  string `json:"RacingNumber"`
	BroadcastName string `json:"BroadcastName"`
	FullName      string `json:"FullName"`
	Tla           string `json:"Tla"`
	Line          int    `json:"Line"`
	TeamName      string `json:"TeamName"`
	TeamColour    string `json:"TeamColour"`
	Reference     string `json:"Reference"`
	HeadshotUrl   string `json:"HeadshotUrl"`
	CountryCode   string `json:"CountryCode"`
}

const lateDeclarationCutoff = 60 * time.Minute

func ParseDriverList(data []byte) (map[string]*DriverInfo, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: DriverList: %w", err)
	}

	var d DriverListDecoder
	for i, rec := range records {
		if err := d.Feed(rec.Elapsed, rec.Payload); err != nil {
			return nil, fmt.Errorf("feeds: DriverList: record %d: %w", i, err)
		}
	}
	return d.Snapshot()
}

type DriverListDecoder struct {
	state     map[string]any
	firstSeen map[string]time.Duration
}

func (d *DriverListDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var patch map[string]any
	if err := json.Unmarshal(decoded, &patch); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	if d.state == nil {
		d.state = make(map[string]any)
		d.firstSeen = make(map[string]time.Duration)
	}
	for drv := range patch {
		if _, ok := d.firstSeen[drv]; !ok {
			d.firstSeen[drv] = elapsed
		}
	}
	wire.MergeInto(d.state, patch)
	return nil
}

func (d *DriverListDecoder) Snapshot() (map[string]*DriverInfo, error) {
	drivers := make(map[string]*DriverInfo, len(d.state))
	for drv, v := range d.state {
		if d.firstSeen[drv] > lateDeclarationCutoff {
			continue
		}

		entry, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("driver %s: entry is %T, want an object", drv, v)
		}

		raw, err := json.Marshal(entry)
		if err != nil {
			return nil, fmt.Errorf("driver %s: %w", drv, err)
		}
		var rd rawDriverInfo
		if err := json.Unmarshal(raw, &rd); err != nil {
			return nil, fmt.Errorf("driver %s: %w", drv, err)
		}

		driverNumber := rd.RacingNumber
		if driverNumber == "" {
			driverNumber = drv
		}

		drivers[drv] = &DriverInfo{
			DriverNumber:  driverNumber,
			Abbreviation:  rd.Tla,
			BroadcastName: rd.BroadcastName,
			FullName:      rd.FullName,
			TeamName:      rd.TeamName,
			TeamColour:    rd.TeamColour,
			Reference:     rd.Reference,
			HeadshotUrl:   rd.HeadshotUrl,
			CountryCode:   rd.CountryCode,
			Line:          rd.Line,
		}
	}

	return drivers, nil
}
