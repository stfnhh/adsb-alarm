package internal

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

type ADSBResponse struct {
	AC []Aircraft `json:"ac"`
}

type Aircraft struct {
	Hex      string  `json:"hex"`
	Category string  `json:"category"`
	Flight   string  `json:"flight"`
	Track    float64 `json:"track"`
	Dir      float64 `json:"dir"`
	Dst      float64 `json:"dst"`
}

type QuietTime struct {
	time.Time
}

func (qt *QuietTime) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))

	// allow empty (optional field)
	if s == "" {
		qt.Time = time.Time{}
		return nil
	}

	t, err := time.Parse("15:04", s)
	if err != nil {
		slog.Error("invalid_quiet_time",
			"value", s,
			"expected_format", "HH:MM",
			"error", err.Error(),
		)
		return fmt.Errorf("quiet time must be HH:MM, got %q", s)
	}

	qt.Time = t

	return nil
}
