package internal

import (
	"log/slog"
	"math"
	"strings"
)

var lastDistance = map[string]float64{}

func isTowardYou(track, dir float64) bool {
	diff := math.Abs(track - dir)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff <= 90
}

func headingDiff(track, dir float64) float64 {
	diff := math.Abs(track - dir)
	if diff > 180 {
		diff = 360 - diff
	}
	return diff
}

func isApproaching(hex string, newDst float64) bool {
	old, exists := lastDistance[hex]
	lastDistance[hex] = newDst
	if !exists {
		return true
	}
	return newDst < old
}

func EvaluateAircraft(cfg Config, ac []Aircraft) []string {
	matches := []string{}

	for _, a := range ac {
		flight := strings.TrimSpace(a.Flight)
		cat := strings.ToUpper(strings.TrimSpace(a.Category))

		// Suppression check
		if IsSuppressed(a.Hex) {
			slog.Info("skip_suppressed",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"reason", "suppressed",
			)
			continue
		}

		// Category filtering
		if !cfg.CategorySet[cat] {
			continue
		}

		// Geometric checks
		diff := headingDiff(a.Track, a.Dir)

		// Aircraft is flying away (opening range)
		if a.Track != 0 && !isTowardYou(a.Track, a.Dir) {
			slog.Info("skip_opening_range",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"track_deg", a.Track,
				"bearing_to_observer_deg", a.Dir,
				"angle_off_bearing_deg", diff,
				"reason", "opening_range",
			)
			continue
		}

		// Aircraft is not reducing distance
		if !isApproaching(a.Hex, a.Dst) {
			slog.Info("skip_range_increasing",
				"hex", a.Hex,
				"flight", flight,
				"category", cat,
				"distance_nm", a.Dst,
				"reason", "range_increasing",
			)
			continue
		}

		// Match — closing range and inbound
		slog.Info("match_closing_range",
			"hex", a.Hex,
			"flight", flight,
			"category", cat,
			"distance_nm", a.Dst,
			"track_deg", a.Track,
			"bearing_to_observer_deg", a.Dir,
			"angle_off_bearing_deg", diff,
			"reason", "closing_range",
		)

		matches = append(matches, a.Hex)
	}

	return matches
}
