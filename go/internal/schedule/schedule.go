package schedule

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
)

// FetchSchedule returns the season schedule, combining the live-timing archive Index with Jolpica's schedule.
//
// The Index supplies session archive paths, circuit keys and F1's event names, but it can be missing meetings (e.g. 2024 rounds 1-9) and its round numbers can be wrong (e.g. 2026 after cancellations). Jolpica supplies the authoritative list of rounds and their numbers. Meetings missing from the Index get derived archive paths (see jolpica.ScheduleSession.PathCandidates).
//
// When the live-timing Index is unavailable, fallbackReason holds its error, so callers can explain why sessions may not load.
func FetchSchedule(ctx context.Context, ltClient *livetiming.Client, jpClient *jolpica.Client, year int) (schedules []jolpica.Schedule, fallbackReason error, err error) {
	season, ltErr := ltClient.FetchSeason(ctx, year)
	jp, jpErr := jpClient.FetchSchedule(ctx, year)

	switch {
	case ltErr != nil && jpErr != nil:
		return nil, nil, fmt.Errorf("schedule: both backends failed for %d: livetiming: %v; jolpica: %w", year, ltErr, jpErr)
	case ltErr != nil:
		return withDerivedPaths(jp), ltErr, nil
	case jpErr != nil:
		return scheduleFromLivetiming(season), nil, nil
	default:
		return merge(scheduleFromLivetiming(season), jp), nil, nil
	}
}

func merge(lt, jp []jolpica.Schedule) []jolpica.Schedule {
	used := make([]bool, len(jp))
	out := make([]jolpica.Schedule, 0, len(jp)+2)

	for _, ev := range lt {
		if ev.EventFormat != "testing" {
			if k := matchJolpica(ev, jp, used); k >= 0 {
				used[k] = true
				ev.RoundNumber = jp[k].RoundNumber
			}
		}
		out = append(out, ev)
	}
	for k, ev := range jp {
		if !used[k] {
			out = append(out, withDerivedPaths([]jolpica.Schedule{ev})[0])
		}
	}

	sort.SliceStable(out, func(a, b int) bool { return eventStart(out[a]).Before(eventStart(out[b])) })
	return out
}

func matchJolpica(ev jolpica.Schedule, jp []jolpica.Schedule, used []bool) int {
	if race, ok := raceStart(ev); ok {
		for k := range jp {
			if jpRace, ok := raceStart(jp[k]); ok && !used[k] && absDuration(race.Sub(jpRace)) <= 24*time.Hour {
				return k
			}
		}
	}
	for k := range jp {
		if !used[k] && strings.EqualFold(jp[k].EventName, ev.EventName) {
			return k
		}
	}
	return -1
}

func raceStart(ev jolpica.Schedule) (time.Time, bool) {
	for _, s := range ev.Sessions {
		if s.SessionType == "Race" {
			return s.StartDateUtc, true
		}
	}
	return time.Time{}, false
}

func eventStart(ev jolpica.Schedule) time.Time {
	if len(ev.Sessions) == 0 {
		return time.Time{}
	}
	return ev.Sessions[0].StartDateUtc
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func withDerivedPaths(events []jolpica.Schedule) []jolpica.Schedule {
	out := make([]jolpica.Schedule, len(events))
	for i, ev := range events {
		ev.EventFormat = "conventional"
		race, hasRace := raceStart(ev)
		sessions := make([]jolpica.ScheduleSession, len(ev.Sessions))
		for j, s := range ev.Sessions {
			if s.SessionType == "Sprint" {
				ev.EventFormat = "sprint"
			}
			if s.Path == "" && hasRace {
				s.PathCandidates = derivePaths(ev.Season, ev.EventName, race, s.SessionType, s.StartDateUtc)
			}
			sessions[j] = s
		}
		ev.Sessions = sessions
		out[i] = ev
	}
	return out
}

func derivePaths(year int, eventName string, raceUtc time.Time, sessionType string, sessionUtc time.Time) []string {
	shifts := [][2]int{{0, 0}, {-1, -1}, {0, -1}, {1, 1}, {0, 1}, {-1, 0}, {1, 0}}
	names := []string{sessionType}
	if year == 2023 && sessionType == "Sprint Qualifying" {
		names = []string{"Sprint Shootout", sessionType}
	}

	event := strings.ReplaceAll(eventName, " ", "_")
	var out []string
	seen := map[string]bool{}
	for _, name := range names {
		for _, sh := range shifts {
			p := fmt.Sprintf("%d/%s_%s/%s_%s/", year,
				raceUtc.AddDate(0, 0, sh[0]).Format("2006-01-02"), event,
				sessionUtc.AddDate(0, 0, sh[1]).Format("2006-01-02"), strings.ReplaceAll(name, " ", "_"))
			if !seen[p] {
				seen[p] = true
				out = append(out, p)
			}
		}
	}
	return out
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
	offset, err := wire.ParseGmtOffset(gmtOffset)
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

func isTestingName(eventName string) bool {
	for _, word := range strings.Fields(strings.ToLower(eventName)) {
		if word == "test" || word == "testing" {
			return true
		}
	}
	return false
}
