package multiviewer

import (
	"context"
	"encoding/json"
	"fmt"
)

type Corner struct {
	Number int
	Letter string
	X      float64
	Y      float64
	Angle  float64
	Length float64
}

type MarshalLight struct {
	Number int
	Letter string
	X      float64
	Y      float64
	Angle  float64
	Length float64
}

type MarshalSector struct {
	Number int
	Letter string
	X      float64
	Y      float64
	Angle  float64
	Length float64
}

type CircuitInfo struct {
	Corners        []Corner
	MarshalLights  []MarshalLight
	MarshalSectors []MarshalSector

	Rotation float64
}

type rawMarker struct {
	Number        int     `json:"number"`
	Letter        string  `json:"letter"`
	Angle         float64 `json:"angle"`
	Length        float64 `json:"length"`
	TrackPosition struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"trackPosition"`
}

type rawCircuitInfo struct {
	Corners        []rawMarker `json:"corners"`
	MarshalLights  []rawMarker `json:"marshalLights"`
	MarshalSectors []rawMarker `json:"marshalSectors"`
	Rotation       float64     `json:"rotation"`
}

func ParseCircuitInfo(data []byte) (CircuitInfo, error) {
	var raw rawCircuitInfo
	if err := json.Unmarshal(data, &raw); err != nil {
		return CircuitInfo{}, fmt.Errorf("multiviewer: CircuitInfo: unmarshal: %w", err)
	}

	return CircuitInfo{
		Corners:        markersToCorners(raw.Corners),
		MarshalLights:  markersToMarshalLights(raw.MarshalLights),
		MarshalSectors: markersToMarshalSectors(raw.MarshalSectors),
		Rotation:       raw.Rotation,
	}, nil
}

func markersToCorners(markers []rawMarker) []Corner {
	out := make([]Corner, 0, len(markers))
	for _, m := range markers {
		out = append(out, Corner{
			Number: m.Number, Letter: m.Letter,
			X: m.TrackPosition.X, Y: m.TrackPosition.Y,
			Angle: m.Angle, Length: m.Length,
		})
	}
	return out
}

func markersToMarshalLights(markers []rawMarker) []MarshalLight {
	out := make([]MarshalLight, 0, len(markers))
	for _, m := range markers {
		out = append(out, MarshalLight{
			Number: m.Number, Letter: m.Letter,
			X: m.TrackPosition.X, Y: m.TrackPosition.Y,
			Angle: m.Angle, Length: m.Length,
		})
	}
	return out
}

func markersToMarshalSectors(markers []rawMarker) []MarshalSector {
	out := make([]MarshalSector, 0, len(markers))
	for _, m := range markers {
		out = append(out, MarshalSector{
			Number: m.Number, Letter: m.Letter,
			X: m.TrackPosition.X, Y: m.TrackPosition.Y,
			Angle: m.Angle, Length: m.Length,
		})
	}
	return out
}

func (c *Client) FetchCircuitInfo(ctx context.Context, circuitKey, year int) (CircuitInfo, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/api/v1/circuits/%d/%d", circuitKey, year))
	if err != nil {
		return CircuitInfo{}, fmt.Errorf("multiviewer: fetch circuit %d/%d: %w", circuitKey, year, err)
	}
	return ParseCircuitInfo(body)
}
