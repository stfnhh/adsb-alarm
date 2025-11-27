package internal

import (
	"sync"
	"time"
)

var (
	suppressedUntil = make(map[string]int64)
	suppressionMux  sync.Mutex
)

// Suppress marks aircraft hex IDs as suppressed until TTL expires.
func Suppress(cfg Config, hexes []string) {
	suppressionMux.Lock()
	defer suppressionMux.Unlock()

	exp := time.Now().Add(cfg.SuppressionTTL).Unix()

	for _, h := range hexes {
		suppressedUntil[h] = exp
	}
}

// IsSuppressed returns true if an aircraft's suppression period is still active.
func IsSuppressed(hex string) bool {
	suppressionMux.Lock()
	defer suppressionMux.Unlock()

	exp, exists := suppressedUntil[hex]
	if !exists {
		return false
	}

	// expired?
	if time.Now().Unix() > exp {
		delete(suppressedUntil, hex)
		return false
	}

	return true
}
