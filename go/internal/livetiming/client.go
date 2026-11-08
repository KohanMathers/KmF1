package livetiming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
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

const settledAfter = 7 * 24 * time.Hour

// StatusError reports an HTTP error status returned by a live-timing host.
type StatusError struct {
	StatusCode int
	URL        string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status %d from %s", e.StatusCode, e.URL)
}

// IsAccessDenied reports whether err contains a 401 or 403 response from a live-timing host.
func IsAccessDenied(err error) bool {
	var se *StatusError
	if !errors.As(err, &se) {
		return false
	}
	return se.StatusCode == http.StatusUnauthorized || se.StatusCode == http.StatusForbidden
}

type Client struct {
	http  *transport.Client
	cache *cache.Cache

	baseURL       string
	mirrorBaseURL string

	authToken string
}

// SetAuthToken sets the F1TV subscription token sent as a bearer token to livetiming.formula1.com. It is never sent to the mirror.
func (c *Client) SetAuthToken(token string) {
	c.authToken = token
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

	body, primaryErr := c.get(ctx, url, c.authToken)
	if primaryErr != nil {
		if IsAccessDenied(primaryErr) {
			if c.authToken == "" {
				primaryErr = fmt.Errorf("%w (the archive requires an F1TV subscription token; set KMF1_F1TV_TOKEN)", primaryErr)
			} else {
				primaryErr = fmt.Errorf("%w (the F1TV token was rejected; it may have expired)", primaryErr)
			}
		}
		var mirrorErr error
		body, mirrorErr = c.get(ctx, c.mirrorBaseURL+path, "")
		if mirrorErr != nil {
			fetchErr := fmt.Errorf("primary and mirror both failed: %w", errors.Join(
				fmt.Errorf("primary: %w", primaryErr),
				fmt.Errorf("mirror: %w", mirrorErr),
			))
			if stale, found, err := c.cache.GetStale(url, parserVersion); err == nil && found {
				return stale, nil
			}
			return nil, fetchErr
		}
	}

	if err := c.cache.Set(url, parserVersion, body, expiryFor(path, time.Now())); err != nil {
		return nil, fmt.Errorf("cache set: %w", err)
	}

	return body, nil
}

var (
	indexPathRe   = regexp.MustCompile(`^/static/(\d{4})/Index\.json$`)
	sessionDateRe = regexp.MustCompile(`^/static/\d{4}/[^/]+/(\d{4}-\d{2}-\d{2})_[^/]+/`)
)

func expiryFor(path string, now time.Time) time.Time {
	if m := indexPathRe.FindStringSubmatch(path); m != nil {
		if year, err := strconv.Atoi(m[1]); err == nil && year < now.Year() {
			return time.Time{}
		}
	}
	if m := sessionDateRe.FindStringSubmatch(path); m != nil {
		if date, err := time.Parse("2006-01-02", m[1]); err == nil && now.Sub(date) > settledAfter {
			return time.Time{}
		}
	}
	return now.Add(defaultTTL)
}

func (c *Client) get(ctx context.Context, url, authToken string) ([]byte, error) {
	headers := map[string]string{"User-Agent": userAgent}
	if authToken != "" {
		headers["Authorization"] = "Bearer " + authToken
	}
	resp, err := c.http.Get(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, &StatusError{StatusCode: resp.StatusCode, URL: url}
	}

	return io.ReadAll(resp.Body)
}
