package jolpica

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func syntheticStatusPage(offset, n, total int) []byte {
	rows := make([]string, n)
	for i := range rows {
		id := offset + i + 1
		rows[i] = fmt.Sprintf(`{"statusId":"%d","count":"1","status":"Status %d"}`, id, id)
	}
	return fmt.Appendf(nil, `{"MRData":{"limit":"%d","offset":"%d","total":"%d","StatusTable":{"Status":[%s]}}}`,
		pageLimit, offset, total, strings.Join(rows, ","))
}

func syntheticDriverPage(offset, n, total int) []byte {
	rows := make([]string, n)
	for i := range rows {
		rows[i] = fmt.Sprintf(`{"driverId":"d%d"}`, offset+i)
	}
	return fmt.Appendf(nil, `{"MRData":{"limit":"%d","offset":"%d","total":"%d","DriverTable":{"Drivers":[%s]}}}`,
		pageLimit, offset, total, strings.Join(rows, ","))
}

func pagedDriverServer(t *testing.T, total int, requests *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		q := r.URL.Query()
		if q.Get("limit") != strconv.Itoa(pageLimit) {
			t.Errorf("limit = %q, want %d", q.Get("limit"), pageLimit)
		}
		offset, err := strconv.Atoi(q.Get("offset"))
		if err != nil {
			t.Errorf("bad offset %q", q.Get("offset"))
		}
		_, _ = w.Write(syntheticDriverPage(offset, min(pageLimit, total-offset), total))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchAllDriverReferences_FollowsPagination(t *testing.T) {
	var requests atomic.Int32
	srv := pagedDriverServer(t, 873, &requests)

	c := newGlobalRefTestClient(t, srv.URL)
	drivers, err := c.FetchAllDriverReferences(context.Background())
	if err != nil {
		t.Fatalf("FetchAllDriverReferences: %v", err)
	}
	if len(drivers) != 873 {
		t.Fatalf("got %d drivers, want 873", len(drivers))
	}
	for i, d := range drivers {
		if want := fmt.Sprintf("d%d", i); d.DriverId != want {
			t.Fatalf("drivers[%d] = %q, want %q (pages out of order or duplicated)", i, d.DriverId, want)
		}
	}
	if got := requests.Load(); got != 9 {
		t.Errorf("made %d requests, want 9", got)
	}
}

func TestFetchAllDriverReferences_CachesMergedList(t *testing.T) {
	var requests atomic.Int32
	srv := pagedDriverServer(t, 250, &requests)

	c := newGlobalRefTestClient(t, srv.URL)
	for range 2 {
		drivers, err := c.FetchAllDriverReferences(context.Background())
		if err != nil {
			t.Fatalf("FetchAllDriverReferences: %v", err)
		}
		if len(drivers) != 250 {
			t.Fatalf("got %d drivers, want 250", len(drivers))
		}
	}
	if got := requests.Load(); got != 3 {
		t.Errorf("made %d requests across two calls, want 3 (second call served from cache)", got)
	}
}

func TestFetchPitStops_FollowsPagination(t *testing.T) {
	const total = 130
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		n := min(pageLimit, total-offset)
		rows := make([]string, n)
		for i := range rows {
			rows[i] = fmt.Sprintf(`{"driverId":"d","lap":"%d","stop":"1","time":"14:00:00","duration":"22.5"}`, offset+i+1)
		}
		_, _ = fmt.Fprintf(w, `{"MRData":{"limit":"%d","offset":"%d","total":"%d","RaceTable":{"Races":[{"PitStops":[%s]}]}}}`,
			pageLimit, offset, total, strings.Join(rows, ","))
	}))
	defer srv.Close()

	c := newGlobalRefTestClient(t, srv.URL)
	stops, err := c.FetchPitStops(context.Background(), 2024, 9)
	if err != nil {
		t.Fatalf("FetchPitStops: %v", err)
	}
	if len(stops) != total {
		t.Fatalf("got %d stops, want %d", len(stops), total)
	}
	if stops[total-1].Lap != total {
		t.Errorf("last stop lap = %d, want %d", stops[total-1].Lap, total)
	}
}
