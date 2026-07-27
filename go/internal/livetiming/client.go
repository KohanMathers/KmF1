package livetiming

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/livetiming/wire"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

const (
	primaryBaseURL = "https://livetiming.formula1.com"
	mirrorBaseURL  = "https://livetiming-mirror.fastf1.dev"
)

const userAgent = "BestHTTP"

const parserVersion = 1

const defaultTTL = 12 * time.Hour

type Client struct {
	http  *transport.Client
	cache *cache.Cache

	baseURL       string
	mirrorBaseURL string
}

func NewClient(t *transport.Client, c *cache.Cache) *Client {
	return &Client{
		http:          t,
		cache:         c,
		baseURL:       primaryBaseURL,
		mirrorBaseURL: mirrorBaseURL,
	}
}

func NewClientWithBaseURL(t *transport.Client, c *cache.Cache, baseURL string) *Client {
	return &Client{http: t, cache: c, baseURL: baseURL, mirrorBaseURL: baseURL}
}

func (c *Client) FetchSeason(ctx context.Context, year int) (*Season, error) {
	body, err := c.fetch(ctx, fmt.Sprintf("/static/%d/Index.json", year))
	if err != nil {
		return nil, fmt.Errorf("livetiming: fetch season %d index: %w", year, err)
	}

	var season Season
	if err := json.Unmarshal(wire.TrimBOM(body), &season); err != nil {
		return nil, fmt.Errorf("livetiming: parse season %d index: %w", year, err)
	}

	return &season, nil
}

func (c *Client) FetchFeed(ctx context.Context, sessionPath, feedFile string) ([]byte, error) {
	body, err := c.fetch(ctx, "/static/"+sessionPath+feedFile)
	if err != nil {
		return nil, fmt.Errorf("livetiming: fetch %s%s: %w", sessionPath, feedFile, err)
	}
	return body, nil
}

func (c *Client) fetch(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL + path

	if body, found, err := c.cache.Get(url, parserVersion); err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	} else if found {
		return body, nil
	}

	body, err := c.get(ctx, url)
	if err != nil {

		body, err = c.get(ctx, c.mirrorBaseURL+path)
		if err != nil {
			return nil, fmt.Errorf("primary and mirror both failed: %w", err)
		}
	}

	if err := c.cache.Set(url, parserVersion, body, time.Now().Add(defaultTTL)); err != nil {
		return nil, fmt.Errorf("cache set: %w", err)
	}

	return body, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	resp, err := c.http.Get(ctx, url, map[string]string{"User-Agent": userAgent})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, url)
	}

	return io.ReadAll(resp.Body)
}
