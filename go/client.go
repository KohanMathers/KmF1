package kmf1

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/jolpica"
	"github.com/KohanMathers/KmF1/go/internal/livetiming"
	"github.com/KohanMathers/KmF1/go/internal/multiviewer"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

// Client is the entry point for the kmf1 package. It holds the shared caches and sub-clients used to fetch live-timing, Jolpica, and MultiViewer data, and should be created with New.
type Client struct {
	livetiming  *livetiming.Client
	jolpica     *jolpica.Client
	multiviewer *multiviewer.Client
	cache       *cache.Cache
	logger      *slog.Logger
}

// Option configures a Client created by New.
type Option func(*Client)

// WithLogger sets the logger the Client uses for diagnostic output. If unset, log messages are discarded.
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) { c.logger = logger }
}

// New creates a Client backed by a persistent cache at cachePath. The cache is created if it does not already exist.
func New(cachePath string, opts ...Option) (*Client, error) {
	c, err := cache.Open(cachePath)
	if err != nil {
		return nil, fmt.Errorf("kmf1: open cache: %w", err)
	}

	t := transport.New()
	client := &Client{
		livetiming:  livetiming.NewClient(t, c),
		jolpica:     jolpica.NewClient(t, c),
		multiviewer: multiviewer.NewClient(t, c),
		cache:       c,
	}
	for _, opt := range opts {
		opt(client)
	}
	return client, nil
}

func (c *Client) log() *slog.Logger {
	if c.logger != nil {
		return c.logger
	}
	return slog.New(slog.DiscardHandler)
}

// Close closes the Client's underlying cache. It should be called when the Client is no longer needed.
func (c *Client) Close() error {
	return c.cache.Close()
}

// Year returns a Season handle for the given year, through which events, standings, and styling data for that season can be fetched.
func (c *Client) Year(year int) *Season {
	return &Season{client: c, year: year}
}

// DriverRefs fetches the full list of driver reference records known to Jolpica, across all seasons.
func (c *Client) DriverRefs(ctx context.Context) ([]jolpica.DriverReference, error) {
	return c.jolpica.FetchAllDriverReferences(ctx)
}

// DriverRef fetches the Jolpica driver reference for a single driver by its Jolpica driver ID (e.g. "verstappen").
func (c *Client) DriverRef(ctx context.Context, driverId string) (jolpica.DriverReference, error) {
	return c.jolpica.FetchDriverReference(ctx, driverId)
}

// Constructors fetches the full list of constructor reference records known to Jolpica, across all seasons.
func (c *Client) Constructors(ctx context.Context) ([]jolpica.ConstructorReference, error) {
	return c.jolpica.FetchAllConstructorReferences(ctx)
}

// Constructor fetches the Jolpica constructor reference for a single constructor by its Jolpica constructor ID (e.g. "red_bull").
func (c *Client) Constructor(ctx context.Context, constructorId string) (jolpica.ConstructorReference, error) {
	return c.jolpica.FetchConstructorReference(ctx, constructorId)
}

// CircuitRefs fetches the full list of circuit reference records known to Jolpica, across all seasons.
func (c *Client) CircuitRefs(ctx context.Context) ([]jolpica.CircuitReference, error) {
	return c.jolpica.FetchAllCircuitReferences(ctx)
}

// CircuitRef fetches the Jolpica circuit reference for a single circuit by its Jolpica circuit ID (e.g. "monza").
func (c *Client) CircuitRef(ctx context.Context, circuitId string) (jolpica.CircuitReference, error) {
	return c.jolpica.FetchCircuitReference(ctx, circuitId)
}

// FinishingStatuses fetches the list of possible session finishing statuses (e.g. "Finished", "Retired") known to Jolpica.
func (c *Client) FinishingStatuses(ctx context.Context) ([]jolpica.FinishingStatusReference, error) {
	return c.jolpica.FetchAllFinishingStatusReferences(ctx)
}
