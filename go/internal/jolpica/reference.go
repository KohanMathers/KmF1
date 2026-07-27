package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type DriverReference struct {
	DriverId        string
	Abbreviation    string
	PermanentNumber *int
	GivenName       string
	FamilyName      string
	DateOfBirth     time.Time
	Nationality     string
}

type ConstructorReference struct {
	ConstructorId string
	Name          string
	Nationality   string
}

type CircuitReference struct {
	CircuitId string
	Name      string
	Locality  string
	Country   string
	Latitude  float64
	Longitude float64
}

type FinishingStatusReference struct {
	StatusId string
	Status   string
	Count    int
}

type rawDriver struct {
	DriverId        string `json:"driverId"`
	PermanentNumber string `json:"permanentNumber"`
	Code            string `json:"code"`
	GivenName       string `json:"givenName"`
	FamilyName      string `json:"familyName"`
	DateOfBirth     string `json:"dateOfBirth"`
	Nationality     string `json:"nationality"`
}

type rawConstructor struct {
	ConstructorId string `json:"constructorId"`
	Name          string `json:"name"`
	Nationality   string `json:"nationality"`
}

type rawCircuit struct {
	CircuitId   string `json:"circuitId"`
	CircuitName string `json:"circuitName"`
	Location    struct {
		Lat      string `json:"lat"`
		Long     string `json:"long"`
		Locality string `json:"locality"`
		Country  string `json:"country"`
	} `json:"Location"`
}

type rawStatus struct {
	StatusId string `json:"statusId"`
	Count    string `json:"count"`
	Status   string `json:"status"`
}

func ParseDriverReferences(data []byte) ([]DriverReference, error) {
	var env struct {
		MRData struct {
			DriverTable struct {
				Drivers []rawDriver `json:"Drivers"`
			} `json:"DriverTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: DriverReference: unmarshal: %w", err)
	}

	drivers := make([]DriverReference, 0, len(env.MRData.DriverTable.Drivers))
	for _, d := range env.MRData.DriverTable.Drivers {
		dob, _ := time.Parse("2006-01-02", d.DateOfBirth)
		var num *int
		if n, err := strconv.Atoi(d.PermanentNumber); err == nil {
			num = &n
		}
		drivers = append(drivers, DriverReference{
			DriverId:        d.DriverId,
			Abbreviation:    d.Code,
			PermanentNumber: num,
			GivenName:       d.GivenName,
			FamilyName:      d.FamilyName,
			DateOfBirth:     dob,
			Nationality:     d.Nationality,
		})
	}
	return drivers, nil
}

func ParseConstructorReferences(data []byte) ([]ConstructorReference, error) {
	var env struct {
		MRData struct {
			ConstructorTable struct {
				Constructors []rawConstructor `json:"Constructors"`
			} `json:"ConstructorTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: ConstructorReference: unmarshal: %w", err)
	}

	constructors := make([]ConstructorReference, 0, len(env.MRData.ConstructorTable.Constructors))
	for _, c := range env.MRData.ConstructorTable.Constructors {
		constructors = append(constructors, ConstructorReference{
			ConstructorId: c.ConstructorId,
			Name:          c.Name,
			Nationality:   c.Nationality,
		})
	}
	return constructors, nil
}

func ParseCircuitReferences(data []byte) ([]CircuitReference, error) {
	var env struct {
		MRData struct {
			CircuitTable struct {
				Circuits []rawCircuit `json:"Circuits"`
			} `json:"CircuitTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: CircuitReference: unmarshal: %w", err)
	}

	circuits := make([]CircuitReference, 0, len(env.MRData.CircuitTable.Circuits))
	for _, c := range env.MRData.CircuitTable.Circuits {
		lat, _ := strconv.ParseFloat(c.Location.Lat, 64)
		long, _ := strconv.ParseFloat(c.Location.Long, 64)
		circuits = append(circuits, CircuitReference{
			CircuitId: c.CircuitId,
			Name:      c.CircuitName,
			Locality:  c.Location.Locality,
			Country:   c.Location.Country,
			Latitude:  lat,
			Longitude: long,
		})
	}
	return circuits, nil
}

func ParseFinishingStatusReferences(data []byte) ([]FinishingStatusReference, error) {
	var env struct {
		MRData struct {
			StatusTable struct {
				Status []rawStatus `json:"Status"`
			} `json:"StatusTable"`
		} `json:"MRData"`
	}
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("jolpica: FinishingStatusReference: unmarshal: %w", err)
	}

	statuses := make([]FinishingStatusReference, 0, len(env.MRData.StatusTable.Status))
	for _, s := range env.MRData.StatusTable.Status {
		count, _ := strconv.Atoi(s.Count)
		statuses = append(statuses, FinishingStatusReference{
			StatusId: s.StatusId,
			Status:   s.Status,
			Count:    count,
		})
	}
	return statuses, nil
}

func (c *Client) FetchDriverReferences(ctx context.Context, year int) ([]DriverReference, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/drivers.json?limit=1000", year))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch drivers %d: %w", year, err)
	}
	return ParseDriverReferences(body)
}

func (c *Client) FetchConstructorReferences(ctx context.Context, year int) ([]ConstructorReference, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/constructors.json?limit=1000", year))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch constructors %d: %w", year, err)
	}
	return ParseConstructorReferences(body)
}

func (c *Client) FetchCircuitReferences(ctx context.Context, year int) ([]CircuitReference, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/circuits.json?limit=1000", year))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch circuits %d: %w", year, err)
	}
	return ParseCircuitReferences(body)
}

func (c *Client) FetchFinishingStatusReferences(ctx context.Context, year int) ([]FinishingStatusReference, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/%d/status.json?limit=1000", year))
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch status %d: %w", year, err)
	}
	return ParseFinishingStatusReferences(body)
}

func (c *Client) FetchAllFinishingStatusReferences(ctx context.Context) ([]FinishingStatusReference, error) {
	body, err := c.fetch(ctx, "/status.json?limit=1000")
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch all statuses: %w", err)
	}
	return ParseFinishingStatusReferences(body)
}

func (c *Client) FetchAllDriverReferences(ctx context.Context) ([]DriverReference, error) {
	body, err := c.fetch(ctx, "/drivers.json?limit=1000")
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch all drivers: %w", err)
	}
	return ParseDriverReferences(body)
}

func (c *Client) FetchDriverReference(ctx context.Context, driverId string) (DriverReference, error) {
	body, err := c.fetch(ctx, "/drivers/"+driverId+".json")
	if err != nil {
		return DriverReference{}, fmt.Errorf("jolpica: fetch driver %s: %w", driverId, err)
	}
	drivers, err := ParseDriverReferences(body)
	if err != nil {
		return DriverReference{}, err
	}
	if len(drivers) == 0 {
		return DriverReference{}, fmt.Errorf("jolpica: no driver found for id %q", driverId)
	}
	return drivers[0], nil
}

func (c *Client) FetchAllConstructorReferences(ctx context.Context) ([]ConstructorReference, error) {
	body, err := c.fetch(ctx, "/constructors.json?limit=1000")
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch all constructors: %w", err)
	}
	return ParseConstructorReferences(body)
}

func (c *Client) FetchConstructorReference(ctx context.Context, constructorId string) (ConstructorReference, error) {
	body, err := c.fetch(ctx, "/constructors/"+constructorId+".json")
	if err != nil {
		return ConstructorReference{}, fmt.Errorf("jolpica: fetch constructor %s: %w", constructorId, err)
	}
	constructors, err := ParseConstructorReferences(body)
	if err != nil {
		return ConstructorReference{}, err
	}
	if len(constructors) == 0 {
		return ConstructorReference{}, fmt.Errorf("jolpica: no constructor found for id %q", constructorId)
	}
	return constructors[0], nil
}

func (c *Client) FetchAllCircuitReferences(ctx context.Context) ([]CircuitReference, error) {
	body, err := c.fetch(ctx, "/circuits.json?limit=1000")
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch all circuits: %w", err)
	}
	return ParseCircuitReferences(body)
}

func (c *Client) FetchCircuitReference(ctx context.Context, circuitId string) (CircuitReference, error) {
	body, err := c.fetch(ctx, "/circuits/"+circuitId+".json")
	if err != nil {
		return CircuitReference{}, fmt.Errorf("jolpica: fetch circuit %s: %w", circuitId, err)
	}
	circuits, err := ParseCircuitReferences(body)
	if err != nil {
		return CircuitReference{}, err
	}
	if len(circuits) == 0 {
		return CircuitReference{}, fmt.Errorf("jolpica: no circuit found for id %q", circuitId)
	}
	return circuits[0], nil
}
