package main

import (
	"log/slog"
	"net/http"
	"time"

  "github.com/stfnhh/adsb-alarm/internal"
)

func main() {
	go func() {
	    http.ListenAndServe("127.0.0.1:8081", nil)
	}()

	cfg := internal.LoadConfig()
	slog.Info("adsb_monitor_start",
		"adsb_url", cfg.ADSBURL,
	)

	slog.Info("quiet_hours_configured",
		"quiet_start_utc", cfg.QuietStart.Format("15:04"),
		"quiet_end_utc", cfg.QuietEnd.Format("15:04"),
	)

	for {
		if internal.InQuietHours(cfg) {
			internal.SleepUntilQuietEnds(cfg)
			continue
		}

		aircraft := internal.FetchAircraft(cfg)
		matches := internal.EvaluateAircraft(cfg, aircraft)

		if len(matches) > 0 {
			internal.TriggerWebhook(cfg)
			internal.Suppress(cfg, matches)
		}

		time.Sleep(cfg.PollInterval)
	}
}
