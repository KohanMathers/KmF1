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
		payloadStart := strings.IndexAny(line, "{[\"")
		if payloadStart < 0 {
			payloadStart = len(line)
		}
		if payloadStart == 0 {
			return nil, fmt.Errorf("wire: line %d has no elapsed-time prefix", i)
		}
		if len(line) < framePrefixLen {
			return nil, fmt.Errorf("wire: line %d shorter than a frame prefix: %q", i, line)
		}

		elapsed, err := parseElapsed(line[:payloadStart])
		if err != nil {
			return nil, fmt.Errorf("wire: line %d: %w", i, err)
		}

		records = append(records, Record{
			Elapsed: elapsed,
			Payload: line[payloadStart:],
		})
	}

	return records, nil
}

func parseElapsed(prefix string) (time.Duration, error) {
	if prefix == "" {
		return 0, fmt.Errorf("malformed elapsed-time prefix %q", prefix)
	}
	var days time.Duration
	clock := prefix
	if dot := strings.IndexByte(prefix, '.'); dot >= 0 {
		dayPart, rest := prefix[:dot], prefix[dot+1:]
		if !strings.Contains(dayPart, ":") {
			d, err := strconv.Atoi(dayPart)
			if err != nil || len(rest) < 8 || rest[2] != ':' || rest[5] != '.' {
				return 0, fmt.Errorf("malformed elapsed-time prefix %q", prefix)
			}
			days = time.Duration(d) * 24 * time.Hour
			clock = rest
		}
	}
	parts := strings.SplitN(clock, ".", 2)
	if len(parts) != 2 || len(parts[1]) < 1 || len(parts[1]) > 9 {
		return 0, fmt.Errorf("malformed elapsed-time prefix %q", prefix)
	}
	hms := strings.Split(parts[0], ":")
	if len(hms) != 3 {
		return 0, fmt.Errorf("malformed elapsed-time prefix %q", prefix)
	}
	h, e1 := strconv.Atoi(hms[0])
	m, e2 := strconv.Atoi(hms[1])
	s, e3 := strconv.Atoi(hms[2])
	frac, e4 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || len(hms[1]) != 2 || len(hms[2]) != 2 {
		return 0, fmt.Errorf("malformed elapsed-time prefix %q", prefix)
	}
	for i := len(parts[1]); i < 9; i++ {
		frac *= 10
	}
	return days + time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second + time.Duration(frac), nil
}

func TrimBOM(data []byte) []byte {
	return bytes.TrimPrefix(data, bom)
}
