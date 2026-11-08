package kmf1

import (
	"fmt"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
)

// Event represents a single race weekend within a Season, obtained via Season.Events or Season.Event.
type Event struct {
	client         *Client
	schedule       jolpica.Schedule
	fallbackReason error
}

func (e *Event) fallbackDetail() string {
	if e.fallbackReason == nil {
		return ""
	}
	return fmt.Sprintf("; live-timing archive error: %v", e.fallbackReason)
}

// Info returns the raw schedule data for the event, including its name, round number, and circuit.
func (e *Event) Info() jolpica.Schedule {
	return e.schedule
}

// Sessions returns the scheduled sessions (practice, qualifying, race, etc.) for the event.
func (e *Event) Sessions() []jolpica.ScheduleSession {
	return e.schedule.Sessions
}

// Session returns a reference to a specific session of the event by its type (e.g. SessionRace, SessionQualifying). The returned SessionRef must be loaded with Load before its data can be read.
func (e *Event) Session(sessionType string) (*SessionRef, error) {
	for _, s := range e.schedule.Sessions {
		if s.SessionType != sessionType {
			continue
		}
		if s.Path == "" {
			return nil, fmt.Errorf("kmf1: no live-timing archive path available for %s %d %s (this event's schedule came from the Jolpica fallback, which has no session paths%s)", e.schedule.EventName, e.schedule.Season, sessionType, e.fallbackDetail())
		}
		return &SessionRef{
			client:      e.client,
			year:        e.schedule.Season,
			round:       e.schedule.RoundNumber,
			sessionType: sessionType,
			path:        s.Path,
		}, nil
	}
	return nil, fmt.Errorf("kmf1: no %q session found for %s %d", sessionType, e.schedule.EventName, e.schedule.Season)
}

// Circuit returns a handle to the circuit geometry data for the event's track, as provided by MultiViewer.
func (e *Event) Circuit() (*Circuit, error) {
	if e.schedule.CircuitKey == 0 {
		return nil, fmt.Errorf("kmf1: no circuit key available for %s %d (this event's schedule came from the Jolpica fallback, which has no circuit key%s)", e.schedule.EventName, e.schedule.Season, e.fallbackDetail())
	}
	return &Circuit{client: e.client, circuitKey: e.schedule.CircuitKey, year: e.schedule.Season}, nil
}

// Standings returns a handle to the championship standings as of this event's round.
func (e *Event) Standings() *Standings {
	return &Standings{client: e.client, year: e.schedule.Season, round: e.schedule.RoundNumber}
}
