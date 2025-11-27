package internal

import (
	"log/slog"
	"time"
)

func InQuietHours(cfg Config) bool {
	now := time.Now().UTC()

	start := time.Date(now.Year(), now.Month(), now.Day(),
		cfg.QuietStart.Hour(), cfg.QuietStart.Minute(), 0, 0, time.UTC)

	end := time.Date(now.Year(), now.Month(), now.Day(),
		cfg.QuietEnd.Hour(), cfg.QuietEnd.Minute(), 0, 0, time.UTC)

	if end.Before(start) {
		return now.After(start) || now.Before(end)
	}
	return now.After(start) && now.Before(end)
}

func SleepUntilQuietEnds(cfg Config) {
	now := time.Now().UTC()
	end := time.Date(
		now.Year(), now.Month(), now.Day(),
		cfg.QuietEnd.Hour(), cfg.QuietEnd.Minute(), 0, 0, time.UTC,
	)

	// If quiet end already passed today, move to tomorrow
	if end.Before(now) {
		end = end.Add(24 * time.Hour)
	}

	d := time.Until(end).Round(time.Second)

	slog.Info("quiet_hours_active",
		"sleep_duration", d.String(),
		"resume_at_utc", end.Format(time.RFC3339),
	)

	time.Sleep(d)

	slog.Info("quiet_hours_end",
		"now_utc", time.Now().UTC().Format(time.RFC3339),
	)
}
