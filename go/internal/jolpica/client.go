package jolpica

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/KohanMathers/KmF1/go/internal/cache"
	"github.com/KohanMathers/KmF1/go/internal/transport"
)

const defaultBaseURL = "https://api.jolpi.ca/ergast/f1"

const userAgent = "KmF1"

const parserVersion = 1

const defaultTTL = 12 * time.Hour

type Client struct {
	http  *transport.Client
	cache *cache.Cache

	baseURL string
}

func NewClient(t *transport.Client, c *cache.Cache) *Client {
	return &Client{http: t, cache: c, baseURL: defaultBaseURL}
}

func NewClientWithBaseURL(t *transport.Client, c *cache.Cache, baseURL string) *Client {
	return &Client{http: t, cache: c, baseURL: baseURL}
}

func (c *Client) fetch(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL + path

	if body, found, err := c.cache.Get(url, parserVersion); err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	} else if found {
		return body, nil
	}

	resp, err := c.http.Get(ctx, url, map[string]string{"User-Agent": userAgent})
	if err != nil {
		return nil, fmt.Errorf("jolpica: fetch %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("jolpica: unexpected status %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("jolpica: read response body: %w", err)
	}

	if err := c.cache.Set(url, parserVersion, body, time.Now().Add(defaultTTL)); err != nil {
		return nil, fmt.Errorf("cache set: %w", err)
	}

	return body, nil
}
