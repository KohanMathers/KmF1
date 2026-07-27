package feeds

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

type MeetingInfo struct {
	Key          int                `json:"Key"`
	Name         string             `json:"Name"`
	OfficialName string             `json:"OfficialName"`
	Location     string             `json:"Location"`
	Country      livetiming.Country `json:"Country"`
	Circuit      livetiming.Circuit `json:"Circuit"`
}

type ArchiveStatus struct {
	Status string `json:"Status"`
}

type SessionInfo struct {
	Meeting   MeetingInfo
	Key       int
	Type      string
	Name      string
	StartDate time.Time
	EndDate   time.Time
	GmtOffset time.Duration
	Path      string

	ArchiveStatus ArchiveStatus
}

type rawSessionInfo struct {
	Meeting       MeetingInfo   `json:"Meeting"`
	ArchiveStatus ArchiveStatus `json:"ArchiveStatus"`
	Key           int           `json:"Key"`
	Type          string        `json:"Type"`
	Name          string        `json:"Name"`
	StartDate     string        `json:"StartDate"`
	EndDate       string        `json:"EndDate"`
	GmtOffset     string        `json:"GmtOffset"`
	Path          string        `json:"Path"`
}

const dateLayout = "2006-01-02T15:04:05"

func ParseSessionInfo(data []byte) (*SessionInfo, error) {
	records, err := wire.SplitFrames(data)
	if err != nil {
		return nil, fmt.Errorf("feeds: SessionInfo: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("feeds: SessionInfo: no records in feed")
	}

	var d SessionInfoDecoder
	if err := d.Feed(records[0].Elapsed, records[0].Payload); err != nil {
		return nil, fmt.Errorf("feeds: SessionInfo: %w", err)
	}
	info, ok := d.Snapshot()
	if !ok {
		return nil, fmt.Errorf("feeds: SessionInfo: no data")
	}
	return info, nil
}

type SessionInfoDecoder struct {
	info *SessionInfo
}

func (d *SessionInfoDecoder) Feed(elapsed time.Duration, payload string) error {
	decoded, err := wire.Decode(payload, false)
	if err != nil {
		return err
	}

	var raw rawSessionInfo
	if err := json.Unmarshal(decoded, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}

	startDate, err := time.Parse(dateLayout, raw.StartDate)
	if err != nil {
		return fmt.Errorf("parse StartDate %q: %w", raw.StartDate, err)
	}
	endDate, err := time.Parse(dateLayout, raw.EndDate)
	if err != nil {
		return fmt.Errorf("parse EndDate %q: %w", raw.EndDate, err)
	}
	gmtOffset, err := parseGmtOffset(raw.GmtOffset)
	if err != nil {
		return fmt.Errorf("parse GmtOffset %q: %w", raw.GmtOffset, err)
	}

	d.info = &SessionInfo{
		Meeting:       raw.Meeting,
		Key:           raw.Key,
		Type:          raw.Type,
		Name:          raw.Name,
		StartDate:     startDate,
		EndDate:       endDate,
		GmtOffset:     gmtOffset,
		Path:          raw.Path,
		ArchiveStatus: raw.ArchiveStatus,
	}
	return nil
}

func (d *SessionInfoDecoder) Snapshot() (*SessionInfo, bool) {
	return d.info, d.info != nil
}

func parseGmtOffset(s string) (time.Duration, error) {
	if len(s) != 8 || s[2] != ':' || s[5] != ':' {
		return 0, fmt.Errorf("malformed GmtOffset %q", s)
	}

	h, err := strconv.Atoi(s[0:2])
	if err != nil {
		return 0, fmt.Errorf("hours in %q: %w", s, err)
	}
	m, err := strconv.Atoi(s[3:5])
	if err != nil {
		return 0, fmt.Errorf("minutes in %q: %w", s, err)
	}
	sec, err := strconv.Atoi(s[6:8])
	if err != nil {
		return 0, fmt.Errorf("seconds in %q: %w", s, err)
	}

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(sec)*time.Second, nil
}
