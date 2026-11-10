package jolpica

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
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

	body, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	if err := c.cache.Set(url, parserVersion, body, time.Now().Add(defaultTTL)); err != nil {
		return nil, fmt.Errorf("cache set: %w", err)
	}

	return body, nil
}

func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	url := c.baseURL + path

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
	return body, nil
}

const pageLimit = 100

const maxPages = 200

type pageMeta struct {
	MRData struct {
		Limit string `json:"limit"`
		Total string `json:"total"`
	} `json:"MRData"`
}

func (c *Client) fetchAllPages(ctx context.Context, path string) ([]json.RawMessage, error) {
	key := fmt.Sprintf("%s%s?limit=%d&offset=*", c.baseURL, path, pageLimit)

	if cached, found, err := c.cache.Get(key, parserVersion); err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	} else if found {
		var pages []json.RawMessage
		if err := json.Unmarshal(cached, &pages); err == nil {
			return pages, nil
		}
	}

	var pages []json.RawMessage
	for offset := 0; ; {
		if len(pages) == maxPages {
			return nil, fmt.Errorf("jolpica: %s: more than %d pages", path, maxPages)
		}
		body, err := c.get(ctx, fmt.Sprintf("%s?limit=%d&offset=%d", path, pageLimit, offset))
		if err != nil {
			return nil, err
		}
		pages = append(pages, body)

		var meta pageMeta
		if err := json.Unmarshal(body, &meta); err != nil {
			return nil, fmt.Errorf("jolpica: %s: unmarshal page metadata: %w", path, err)
		}
		total, _ := strconv.Atoi(meta.MRData.Total)
		limit, _ := strconv.Atoi(meta.MRData.Limit)
		if limit <= 0 || offset+limit >= total {
			break
		}
		offset += limit
	}

	merged, err := json.Marshal(pages)
	if err != nil {
		return nil, fmt.Errorf("jolpica: %s: encode pages: %w", path, err)
	}
	if err := c.cache.Set(key, parserVersion, merged, time.Now().Add(defaultTTL)); err != nil {
		return nil, fmt.Errorf("cache set: %w", err)
	}
	return pages, nil
}

func fetchList[T any](ctx context.Context, c *Client, path string, parse func([]byte) ([]T, error)) ([]T, error) {
	pages, err := c.fetchAllPages(ctx, path)
	if err != nil {
		return nil, err
	}
	rows := make([]T, 0)
	for _, page := range pages {
		pageRows, err := parse(page)
		if err != nil {
			return nil, err
		}
		rows = append(rows, pageRows...)
	}
	return rows, nil
}
