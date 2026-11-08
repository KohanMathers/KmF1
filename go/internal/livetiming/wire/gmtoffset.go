package wire

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseGmtOffset parses a live-timing GmtOffset such as "02:00:00" or "-06:00:00". The sign applies to the whole duration, so "-03:30:00" is -3h30m.
func ParseGmtOffset(s string) (time.Duration, error) {
	body, negative := s, false
	switch {
	case strings.HasPrefix(s, "-"):
		body, negative = s[1:], true
	case strings.HasPrefix(s, "+"):
		body = s[1:]
	}

	parts := strings.Split(body, ":")
	if len(parts) != 3 {
		return 0, fmt.Errorf("malformed GmtOffset %q", s)
	}
	var fields [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || len(p) != 2 {
			return 0, fmt.Errorf("malformed GmtOffset %q", s)
		}
		fields[i] = n
	}

	d := time.Duration(fields[0])*time.Hour + time.Duration(fields[1])*time.Minute + time.Duration(fields[2])*time.Second
	if negative {
		d = -d
	}
	return d, nil
}
