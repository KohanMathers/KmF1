package kmf1

import (
	"context"

	"github.com/KohanMathers/KmF1/go/internal/jolpica"
)

// Standings is a handle to the championship standings as of a particular round in a season, obtained via Event.Standings.
type Standings struct {
	client *Client
	year   int
	round  int
}

// Drivers returns the driver championship standings.
func (s *Standings) Drivers(ctx context.Context) ([]jolpica.DriverStanding, error) {
	return s.client.jolpica.FetchDriverStandings(ctx, s.year, s.round)
}

// Constructors returns the constructor championship standings.
func (s *Standings) Constructors(ctx context.Context) ([]jolpica.ConstructorStanding, error) {
	return s.client.jolpica.FetchConstructorStandings(ctx, s.year, s.round)
}
