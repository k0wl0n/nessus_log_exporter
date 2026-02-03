package main

import (
	"strings"

	"github.com/go-kit/log"
)

// backend.log uses Apache-style combined log format with severity tags
// Format: [DD/Mon/YYYY:HH:MM:SS +0000] [severity] [component] message
// Tags: [info] [warn] [error] [trace] [performance]

func parseBackendLog(path string, state *exporterState, logger log.Logger) error {
	return tailLines(path, state, func(line string) {
		state.backendLines++

		// ── severity detection (tag-based) ─────────────────────
		if containsTag(line, "error") {
			state.errors["backend"]++
		}
		if containsTag(line, "warn") {
			state.warnings["backend"]++
		}

		// backend.log can also mention scan events in some Nessus
		// builds; we still rely on nessusd.messages for the primary
		// scan counters to avoid double-counting.
	})
}

// containsTag checks for [TAG] anywhere in the line (case-insensitive).
func containsTag(line, tag string) bool {
	needle := "[" + tag + "]"
	return strings.Contains(strings.ToLower(line), needle)
}
