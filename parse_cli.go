package main

import (
	"regexp"
	"strings"

	"github.com/go-kit/log"
)

// nessuscli.log format: [DD/Mon/YYYY:HH:MM:SS +0000] message
// May include [severity] [component] tags

// nessuscli invoked with: COMMAND ...
var reCommandInvoked = regexp.MustCompile(`nessuscli invoked with:\s+(\S+)`)

// Successfully linked to sensor.cloud.tenable.com:443
var reLinked = regexp.MustCompile(`(?i)Successfully\s+linked\s+to`)

// unlinked
var reUnlinked = regexp.MustCompile(`(?i)unlinked`)

func parseCLILog(path string, state *exporterState, logger log.Logger) error {
	return tailLines(path, state, func(line string) {
		state.cliLines++

		// ── command extraction ─────────────────────────────────
		if m := reCommandInvoked.FindStringSubmatch(line); m != nil {
			cmd := strings.ToLower(m[1])
			state.cliCommands[cmd]++
		}

		// ── link state transitions ─────────────────────────────
		if reLinked.MatchString(line) {
			state.isAgentLinked = true
		}
		if reUnlinked.MatchString(line) {
			state.isAgentLinked = false
		}

		// ── error / warn ───────────────────────────────────────
		lower := strings.ToLower(line)
		if strings.Contains(lower, "[error]") {
			state.errors["cli"]++
		}
		if strings.Contains(lower, "[warn]") {
			state.warnings["cli"]++
		}
	})
}
