package firewalla

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Timestamp is a Unix epoch timestamp. The MSP API serializes timestamps as
// integer or floating-point numbers of seconds (e.g. 1778413842 or
// 1778413842.883). Timestamp.UnmarshalJSON accepts both forms.
//
// The decoder parses the JSON token as decimal text rather than going
// through float64, so sub-second precision is preserved exactly down to
// nanoseconds.
//
// A JSON null decodes to the zero Timestamp; the literal 0 decodes to
// 1970-01-01T00:00:00Z (a real time.Time, not the zero value).
type Timestamp struct {
	time.Time
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || string(data) == "null" {
		t.Time = time.Time{}
		return nil
	}
	raw := string(data)
	// Strip JSON string quotes if present.
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
		if raw == "" {
			t.Time = time.Time{}
			return nil
		}
		// If the quoted value isn't a decimal number, try RFC3339.
		if !looksLikeNumber(raw) {
			parsed, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				return fmt.Errorf("firewalla: timestamp: %w", err)
			}
			t.Time = parsed.UTC()
			return nil
		}
	}
	parsed, err := parseEpochDecimal(raw)
	if err != nil {
		return fmt.Errorf("firewalla: timestamp: %w", err)
	}
	t.Time = parsed
	return nil
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("0"), nil
	}
	sec := t.Unix()
	nsec := t.Nanosecond()
	if nsec == 0 {
		return []byte(strconv.FormatInt(sec, 10)), nil
	}
	frac := strconv.FormatInt(int64(nsec), 10)
	// Pad to 9 digits then trim trailing zeros for a compact representation.
	if len(frac) < 9 {
		frac = strings.Repeat("0", 9-len(frac)) + frac
	}
	frac = strings.TrimRight(frac, "0")
	return []byte(strconv.FormatInt(sec, 10) + "." + frac), nil
}

// parseEpochDecimal parses `<sec>` or `<sec>.<frac>` into a UTC time.Time.
// Negative epochs are supported.
func parseEpochDecimal(s string) (time.Time, error) {
	secStr, fracStr, hasFrac := strings.Cut(s, ".")
	sec, err := strconv.ParseInt(secStr, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	var nsec int64
	if hasFrac && fracStr != "" {
		if len(fracStr) > 9 {
			fracStr = fracStr[:9]
		} else if len(fracStr) < 9 {
			fracStr = fracStr + strings.Repeat("0", 9-len(fracStr))
		}
		nsec, err = strconv.ParseInt(fracStr, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		if sec < 0 {
			nsec = -nsec
		}
	}
	return time.Unix(sec, nsec).UTC(), nil
}

func looksLikeNumber(s string) bool {
	if s == "" {
		return false
	}
	for i, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c == '.' && i != 0:
		case c == '-' && i == 0:
		case c == '+' && i == 0:
		default:
			return false
		}
	}
	return true
}
