package kmf1

import (
	"context"

	"github.com/KohanMathers/KmF1/go/internal/multiviewer"
)

// Circuit is a handle to MultiViewer's circuit geometry data for a single track and year, obtained via Event.Circuit.
type Circuit struct {
	client     *Client
	circuitKey int
	year       int
}

// Corners returns the track's corner positions and numbering.
func (c *Circuit) Corners(ctx context.Context) ([]multiviewer.Corner, error) {
	info, err := c.client.multiviewer.FetchCircuitInfo(ctx, c.circuitKey, c.year)
	if err != nil {
		return nil, err
	}
	return info.Corners, nil
}

// MarshalLights returns the positions of the track's marshal light panels.
func (c *Circuit) MarshalLights(ctx context.Context) ([]multiviewer.MarshalLight, error) {
	info, err := c.client.multiviewer.FetchCircuitInfo(ctx, c.circuitKey, c.year)
	if err != nil {
		return nil, err
	}
	return info.MarshalLights, nil
}

// MarshalSectors returns the track's marshal sector boundaries.
func (c *Circuit) MarshalSectors(ctx context.Context) ([]multiviewer.MarshalSector, error) {
	info, err := c.client.multiviewer.FetchCircuitInfo(ctx, c.circuitKey, c.year)
	if err != nil {
		return nil, err
	}
	return info.MarshalSectors, nil
}
