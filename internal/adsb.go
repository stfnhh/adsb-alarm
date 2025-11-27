package internal

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func FetchAircraft(cfg Config) []Aircraft {
	client := &http.Client{Timeout: 10 * time.Second}

	req, _ := http.NewRequest("GET", cfg.ADSBURL, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("adsb_request_failed",
			"error", err,
			"url", cfg.ADSBURL,
		)
		return nil
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("response_body_close_failed", "error", cerr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		slog.Error("adsb_bad_status_code",
			"status_code", resp.StatusCode,
			"url", cfg.ADSBURL,
		)
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("adsb_read_failed",
			"error", err,
		)
		return nil
	}

	var parsed ADSBResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		slog.Error("adsb_json_decode_failed",
			"error", err,
			"raw_body", string(body), // includes raw text for debugging
		)
		return nil
	}

	return parsed.AC
}
