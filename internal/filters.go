package internal

import (
	"log/slog"
	"math"
	"strings"
	"time"
)

type distanceState struct {
	dst float64
	ts  time.Time
}

var lastDistance = map[string]distanceState{}

const (
	distanceTTL = 1 * time.Hour
	distEps     = 0.02 // ~120 ft jitter tolerance
)

// Remove stale aircraft state
func pruneDistanceState(now time.Time) {
	for hex, s := range lastDistance {
		if now.Sub(s.ts) > distanceTTL {
			delete(lastDistance, hex)
		}
	}
}

func isGeometricallyInbound(diff float64) bool {
	// Reject anything that is mostly lateral or outbound
	return diff < 90
}

func headingDiff(track, dir float64) float64 {
	diff := math.Abs(track - dir)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

func EvaluateAircraft(cfg Config, ac []Aircraft) []string {
	matches := []string{}
	now := time.Now()

	// Bound memory
	pruneDistanceState(now)

	for _, a := range ac {
		flight := strings.TrimSpace(a.Flight)
		cat := strings.ToUpper(strings.TrimSpace(a.Category))

		// Suppression
		if IsSuppressed(a.Hex) {
			slog.Info("skip_suppressed",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"reason", "suppressed",
			)
			continue
		}

		// Category filter
		if !cfg.CategorySet[cat] {
			continue
		}

		// Geometry
		diff := headingDiff(a.Track, a.Dir)

		// Distance state
		state, exists := lastDistance[a.Hex]

		// Update state AFTER reading
		lastDistance[a.Hex] = distanceState{
			dst: a.Dst,
			ts:  now,
		}

		// Must have at least two samples
		if !exists {
			slog.Info("skip_first_sample",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"distance_nm", a.Dst,
				"reason", "first_sample",
			)
			continue
		}

		deltaDst := state.dst - a.Dst
		deltaT := now.Sub(state.ts).Seconds()

		// Time sanity
		if deltaT <= 0 {
			slog.Info("skip_invalid_time_delta",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"reason", "invalid_time_delta",
			)
			continue
		}

		// Must be meaningfully closing
		if deltaDst <= distEps {
			slog.Info("skip_range_not_decreasing",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"distance_nm", a.Dst,
				"previous_distance_nm", state.dst,
				"reason", "range_not_decreasing",
			)
			continue
		}

		// // Geometry guard — fixes outbound / lateral false positives
		// if diff >= 90 {
		// 	slog.Info("skip_geometrically_outbound",
		// 		"hex", a.Hex,
		// 		"flight", flight,
		// 		"category", cat,
		// 		"angle_off_bearing_deg", diff,
		// 		"reason", "geometry_reject",
		// 	)
		// 	continue
		// }

		// Match
		slog.Info("match_closing_range",
			"hex", a.Hex,
			"flight", flight,
			"category", cat,
			"distance_nm", a.Dst,
			"track_deg", a.Track,
			"bearing_to_observer_deg", a.Dir,
			"angle_off_bearing_deg", diff,
			"delta_distance_nm", deltaDst,
			"delta_time_sec", deltaT,
			"reason", "closing_range",
		)

		matches = append(matches, a.Hex)
	}

	return matches
}
