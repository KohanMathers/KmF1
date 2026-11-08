package schedule

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming"
)

// FetchSchedule returns the season schedule from the live-timing archive, falling back to Jolpica if the archive is unavailable. When the fallback is used, fallbackReason holds the live-timing error so callers can explain why sessions have no archive paths.
func FetchSchedule(ctx context.Context, ltClient *livetiming.Client, jpClient *jolpica.Client, year int) (schedules []jolpica.Schedule, fallbackReason error, err error) {
	season, ltErr := ltClient.FetchSeason(ctx, year)
	if ltErr == nil {
		return scheduleFromLivetiming(season), nil, nil
	}

	schedules, jpErr := jpClient.FetchSchedule(ctx, year)
	if jpErr != nil {
		return nil, nil, fmt.Errorf("schedule: both backends failed for %d: livetiming: %v; jolpica: %w", year, ltErr, jpErr)
	}
	return schedules, ltErr, nil
}

func scheduleFromLivetiming(season *livetiming.Season) []jolpica.Schedule {
	out := make([]jolpica.Schedule, 0, len(season.Meetings))
	for _, m := range season.Meetings {
		format := deriveEventFormat(m.Name, m.Sessions)
		round := m.Number
		if format == "testing" {
			round = 0
		}
		s := jolpica.Schedule{
			Season:            season.Year,
			RoundNumber:       round,
			EventName:         m.Name,
			OfficialEventName: m.OfficialName,
			Country:           m.Country.Name,
			Location:          m.Location,
			EventFormat:       format,
			CircuitKey:        m.Circuit.Key,
		}
		for _, sess := range m.Sessions {
			local, utc, err := parseSessionTimes(sess.StartDate, sess.GmtOffset)
			if err != nil {
				continue
			}
			s.Sessions = append(s.Sessions, jolpica.ScheduleSession{
				SessionType:  sess.Name,
				StartDate:    local,
				StartDateUtc: utc,
				Path:         sess.Path,
			})
		}
		out = append(out, s)
	}
	return out
}

func deriveEventFormat(eventName string, sessions []livetiming.Session) string {
	if isTestingName(eventName) {
		return "testing"
	}
	for _, s := range sessions {
		if s.Name == "Sprint" {
			return "sprint"
		}
	}
	return "conventional"
}

func parseSessionTimes(startDate, gmtOffset string) (local, utc time.Time, err error) {
	offset, err := parseGmtOffset(gmtOffset)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	loc := time.FixedZone("", int(offset.Seconds()))
	local, err = time.ParseInLocation("2006-01-02T15:04:05", startDate, loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return local, local.UTC(), nil
}

func parseGmtOffset(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("unrecognized GmtOffset format %q", s)
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}
	sec, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0, err
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec)*time.Second, nil
}

func isTestingName(eventName string) bool {
	for _, word := range strings.Fields(strings.ToLower(eventName)) {
		if word == "test" || word == "testing" {
			return true
		}
	}
	return false
}
