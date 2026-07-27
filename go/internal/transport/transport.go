package transport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const requestsPerHour = 500

const burst = 25

const defaultTimeout = 30 * time.Second

const maxRetries = 2

var retryBaseDelay = 200 * time.Millisecond

type Client struct {
	http *http.Client

	requestsPerHour int
	burst           int

	mu       sync.Mutex
	limiters map[string]*rate.Limiter
}

func New() *Client {
	return NewWithRate(requestsPerHour, burst)
}

func NewWithRate(requestsPerHour, burst int) *Client {
	return &Client{
		http:            &http.Client{Timeout: defaultTimeout},
		requestsPerHour: requestsPerHour,
		burst:           burst,
		limiters:        make(map[string]*rate.Limiter),
	}
}

func (c *Client) Get(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error) {
	host, err := hostOf(rawURL)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := retryBaseDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("transport: request %s: %w", rawURL, ctx.Err())
			case <-time.After(delay):
			}
		}

		if err := c.limiterFor(host).Wait(ctx); err != nil {
			return nil, fmt.Errorf("transport: rate limit wait for %s: %w", host, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("transport: build request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, doErr := c.http.Do(req)
		if doErr == nil {
			return resp, nil
		}
		lastErr = doErr

		if !isRetryable(doErr) {
			break
		}
	}

	return nil, fmt.Errorf("transport: request %s: %w", rawURL, lastErr)
}

func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return true
	}

	return false
}

func (c *Client) limiterFor(host string) *rate.Limiter {
	c.mu.Lock()
	defer c.mu.Unlock()

	l, ok := c.limiters[host]
	if !ok {
		l = rate.NewLimiter(rate.Every(time.Hour/time.Duration(c.requestsPerHour)), c.burst)
		c.limiters[host] = l
	}
	return l
}

func hostOf(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("transport: parse url %q: %w", rawURL, err)
	}
	return u.Host, nil
}
