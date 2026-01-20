package internal

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	APIKey     string `envconfig:"API_KEY" required:"true"`
	Latitude   string `envconfig:"LATITUDE" required:"true"`
	Longitude  string `envconfig:"LONGITUDE" required:"true"`
	Distance   string `envconfig:"DISTANCE" required:"true"`
	WebhookURL string `envconfig:"WEBHOOK_URL" required:"true"`

	Categories []string `envconfig:"CATEGORIES"`

	PollInterval   time.Duration `envconfig:"POLL_INTERVAL" default:"30s"`
	SuppressionTTL time.Duration `envconfig:"SUPPRESSION_TTL" default:"30m"`

	QuietStart QuietTime `envconfig:"QUIET_START"`
	QuietEnd   QuietTime `envconfig:"QUIET_END"`

	ADSBURL     string          `ignored:"true"`
	CategorySet map[string]bool `ignored:"true"`
}

func loadEnvFile(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// don't override existing env
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func loadSecretEnv(key string, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	os.Setenv(key, strings.TrimSpace(string(data)))
}

func LoadConfig() Config {
	// 1. Load config file
	loadEnvFile("/run/configs/app.env")

	// 2. Load secret
	loadSecretEnv("API_KEY", "/run/secrets/api_key")

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		slog.Error("config_error",
			"error", err.Error(),
		)
		os.Exit(1)
	}

	cfg.ADSBURL = fmt.Sprintf(
		"https://opendata.adsb.fi/api/v3/lat/%s/lon/%s/dist/%s",
		cfg.Latitude, cfg.Longitude, cfg.Distance,
	)

	cfg.CategorySet = map[string]bool{}
	for _, c := range cfg.Categories {
		c = strings.ToUpper(strings.TrimSpace(c))
		if c != "" {
			cfg.CategorySet[c] = true
		}
	}

	return cfg
}

