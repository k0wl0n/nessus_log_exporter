package main

import (
	"strings"

	"github.com/go-kit/log"
)

// nessusd.dump format is highly structured with key-value pairs
// Format: [timestamp][PID.TID][key=val]...[severity=LEVEL] : message
// Common operations: op=sync, op=_defrag, op=_map_lowmem
// When nasl_log_type = normal the file may be empty or very sparse.

func parseDumpLog(path string, state *exporterState, logger log.Logger) error {
	return tailLines(path, state, func(line string) {
		state.dumpLines++

		// Check for [severity=ERROR] or [severity=WARN] in structured format
		if strings.Contains(line, "[severity=ERROR]") {
			state.errors["dump"]++
			return
		}
		if strings.Contains(line, "[severity=WARN]") {
			state.warnings["dump"]++
			return
		}

		// Fallback: check message text for error/warn patterns
		lower := strings.ToLower(line)
		if strings.Contains(lower, ": error") || strings.Contains(lower, "[error]") {
			state.errors["dump"]++
		}
		if strings.Contains(lower, ": warn") || strings.Contains(lower, "[warn]") {
			state.warnings["dump"]++
		}
	})
}
