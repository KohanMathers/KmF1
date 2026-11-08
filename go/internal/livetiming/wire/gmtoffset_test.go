package wire

import (
	"testing"
	"time"
)

func TestParseGmtOffset(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"02:00:00", 2 * time.Hour},
		{"+05:30:00", 5*time.Hour + 30*time.Minute},
		{"00:00:00", 0},
		{"-06:00:00", -6 * time.Hour},
		{"-03:30:00", -(3*time.Hour + 30*time.Minute)},
	}
	for _, c := range cases {
		got, err := ParseGmtOffset(c.in)
		if err != nil {
			t.Errorf("ParseGmtOffset(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseGmtOffset(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseGmtOffset_RejectsMalformed(t *testing.T) {
	for _, in := range []string{"", "bad", "2:00:00", "02:00", "--06:00:00", "-0a:00:00", "02:-1:00"} {
		if _, err := ParseGmtOffset(in); err == nil {
			t.Errorf("ParseGmtOffset(%q): expected an error", in)
		}
	}
}
