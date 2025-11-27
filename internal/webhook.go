package internal

import (
	"log/slog"
	"net/http"
	"time"
)

func TriggerWebhook(cfg Config) {
	client := &http.Client{Timeout: 10 * time.Second}

	req, _ := http.NewRequest("POST", cfg.WebhookURL, nil)
	req.Header.Set("X-API-KEY", cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("webhook_request_failed",
			"url", cfg.WebhookURL,
			"error", err.Error(),
		)
		return
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Warn("webhook_resp_close_failed",
				"error", cerr.Error(),
			)
		}
	}()

	slog.Info("webhook_triggered",
		"url", cfg.WebhookURL,
		"status_code", resp.StatusCode,
	)
}
