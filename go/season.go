package kmf1

import (
	"context"
	"fmt"
	"strings"

	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/schedule"
	"github.com/KohanMathers/KmF1/go/internal/styling"
)

// Season is a handle to a single Formula 1 season, obtained via Client.Year. It is used to look up events, standings, and styling data for that year.
type Season struct {
	client *Client
	year   int
}

// Events fetches the full race calendar for the season, in schedule order.
func (s *Season) Events(ctx context.Context) ([]Event, error) {
	schedules, fallbackReason, err := schedule.FetchSchedule(ctx, s.client.livetiming, s.client.jolpica, s.year)
	if err != nil {
		return nil, fmt.Errorf("kmf1: fetch events for %d: %w", s.year, err)
	}
	if fallbackReason != nil {
		s.client.log().Warn("live-timing archive unavailable, using Jolpica schedule without session paths",
			"year", s.year, "access_denied", livetiming.IsAccessDenied(fallbackReason), "error", fallbackReason)
	}

	events := make([]Event, len(schedules))
	for i, sc := range schedules {
		events[i] = Event{client: s.client, schedule: sc, fallbackReason: fallbackReason}
	}
	return events, nil
}

// Event finds a single event in the season by round number (int) or event name (string, case-insensitive). Testing events have round number 0.
func (s *Season) Event(ctx context.Context, identifier any) (*Event, error) {
	events, err := s.Events(ctx)
	if err != nil {
		return nil, err
	}

	switch id := identifier.(type) {
	case int:
		for i := range events {
			if events[i].schedule.RoundNumber == id {
				return &events[i], nil
			}
		}
		return nil, fmt.Errorf("kmf1: no event found for round %d in %d", id, s.year)
	case string:
		for i := range events {
			if strings.EqualFold(events[i].schedule.EventName, id) {
				return &events[i], nil
			}
		}
		return nil, fmt.Errorf("kmf1: no event found matching %q in %d", id, s.year)
	default:
		return nil, fmt.Errorf("kmf1: Event identifier must be an int (round number) or string (event name), got %T", identifier)
	}
}

// TeamStyle returns the styling (colors, name) for a single team in this season, matched by team name or ID. It reports false if the season has no known styling data or the team isn't found.
func (s *Season) TeamStyle(identifier string) (styling.TeamStyle, bool) {
	constants, ok := styling.ForYear(s.year)
	if !ok {
		return styling.TeamStyle{}, false
	}
	return constants.Team(identifier)
}

// TeamStyles returns the styling for every team in this season, or nil if the season has no known styling data.
func (s *Season) TeamStyles() []styling.TeamStyle {
	constants, ok := styling.ForYear(s.year)
	if !ok {
		return nil
	}
	return constants.Teams()
}

// CompoundColor returns the hex color associated with a tyre compound (e.g. "SOFT") in this season. It reports false if the season has no known styling data or the compound isn't found.
func (s *Season) CompoundColor(compound string) (string, bool) {
	constants, ok := styling.ForYear(s.year)
	if !ok {
		return "", false
	}
	return constants.CompoundColor(compound)
}

// Compounds returns the names of all tyre compounds with known styling in this season, or nil if the season has no known styling data.
func (s *Season) Compounds() []string {
	constants, ok := styling.ForYear(s.year)
	if !ok {
		return nil
	}
	return constants.Compounds()
}
