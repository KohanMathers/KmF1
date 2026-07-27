package wire

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var bom = []byte{0xEF, 0xBB, 0xBF}

const framePrefixLen = len("00:00:00.000")

type Record struct {
	Elapsed time.Duration
	Payload string
}

func SplitFrames(data []byte) ([]Record, error) {
	data = bytes.TrimPrefix(data, bom)

	lines := strings.Split(string(data), "\r\n")
	records := make([]Record, 0, len(lines))

	for i, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < framePrefixLen {
			return nil, fmt.Errorf("wire: line %d shorter than a frame prefix: %q", i, line)
		}

		elapsed, err := parseElapsed(line[:framePrefixLen])
		if err != nil {
			return nil, fmt.Errorf("wire: line %d: %w", i, err)
		}

		records = append(records, Record{
			Elapsed: elapsed,
			Payload: line[framePrefixLen:],
		})
	}

	return records, nil
}

func parseElapsed(prefix string) (time.Duration, error) {
	if len(prefix) != framePrefixLen || prefix[2] != ':' || prefix[5] != ':' || prefix[8] != '.' {
		return 0, fmt.Errorf("malformed elapsed-time prefix %q", prefix)
	}

	h, err := strconv.Atoi(prefix[0:2])
	if err != nil {
		return 0, fmt.Errorf("hours in %q: %w", prefix, err)
	}
	m, err := strconv.Atoi(prefix[3:5])
	if err != nil {
		return 0, fmt.Errorf("minutes in %q: %w", prefix, err)
	}
	s, err := strconv.Atoi(prefix[6:8])
	if err != nil {
		return 0, fmt.Errorf("seconds in %q: %w", prefix, err)
	}
	ms, err := strconv.Atoi(prefix[9:12])
	if err != nil {
		return 0, fmt.Errorf("milliseconds in %q: %w", prefix, err)
	}

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(s)*time.Second +
		time.Duration(ms)*time.Millisecond, nil
}

func TrimBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, bom)
}
