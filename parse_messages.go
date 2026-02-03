package main

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

// ── compiled regexes (package-level, compiled once) ─────────────────────────

// [Thu Mar  8 13:26:03 2018][25606.130][JOB_NAME=test_scan][JOB_UUID=…] User admin starts a new scan (…)
var reScanStarted = regexp.MustCompile(
	`\[JOB_NAME=([^\]]+)\].*?(?:User|user)\s+(\S+)\s+starts\s+a\s+new\s+scan`)

// user admin : testing test-linux-host01 (192.168.56.12) [131]
var reHostTesting = regexp.MustCompile(
	`\[JOB_NAME=([^\]]+)\].*?testing\s+(\S+)\s+\(`)

// Finished testing test-linux-host01. Time : 77.58 secs
var reHostFinished = regexp.MustCompile(
	`\[JOB_NAME=([^\]]+)\].*?Finished\s+testing\s+(\S+)\.\s+Time\s*:\s*([\d.]+)\s+secs`)

// Scan done: 1 hosts up
var reScanDone = regexp.MustCompile(
	`\[JOB_NAME=([^\]]+)\].*?Scan\s+done:\s+(\d+)\s+hosts\s+up`)

// test complete  →  scan completed
var reScanComplete = regexp.MustCompile(
	`\[JOB_NAME=([^\]]+)\].*?test\s+complete`)

// nessusd 8.x.x (build Mxxxxx) started
var reVersionStarted = regexp.MustCompile(
	`nessusd\s+([\d.]+)\s+\(build\s+(\S+)\)\s+started`)

// launching <plugin>.nasl against <host>
var rePluginLaunch = regexp.MustCompile(
	`\[JOB_NAME=([^\]]+)\].*?launching\s+\S+\.nasl\s+against\s+`)

// ── parser ────────────────────────────────────────────────────────────────

func parseMessagesLog(path string, state *exporterState, logger log.Logger) error {
	return tailLines(path, state, func(line string) {
		state.messagesLines++

		// ── version / startup ──────────────────────────────────
		if m := reVersionStarted.FindStringSubmatch(line); m != nil {
			state.daemonVersion = m[1]
			state.daemonBuild = m[2]
			state.isRunning = true
			level.Debug(logger).Log("msg", "nessusd started", "version", m[1])
			return
		}

		// ── scan started ───────────────────────────────────────
		if m := reScanStarted.FindStringSubmatch(line); m != nil {
			key := scanUserKey{scanName: m[1], user: m[2]}
			state.scanStarted[key]++
			return
		}

		// ── host being tested ──────────────────────────────────
		if m := reHostTesting.FindStringSubmatch(line); m != nil {
			state.hostsTested++
			return
		}

		// ── host finished (captures duration) ──────────────────
		if m := reHostFinished.FindStringSubmatch(line); m != nil {
			dur, err := strconv.ParseFloat(m[3], 64)
			if err == nil {
				key := scanHostKey{scanName: m[1], host: m[2]}
				state.lastScanDuration[key] = dur
			}
			return
		}

		// ── scan done → hosts up ───────────────────────────────
		if m := reScanDone.FindStringSubmatch(line); m != nil {
			up, err := strconv.Atoi(m[2])
			if err == nil {
				state.hostsUp += up
			}
			return
		}

		// ── scan completed ─────────────────────────────────────
		if m := reScanComplete.FindStringSubmatch(line); m != nil {
			state.scanCompleted[m[1]]++
			return
		}

		// ── plugin launched ────────────────────────────────────
		if m := rePluginLaunch.FindStringSubmatch(line); m != nil {
			state.pluginsLaunched[m[1]]++
			return
		}

		// ── error / warning keywords (simple heuristic) ───────
		lower := strings.ToLower(line)
		if strings.Contains(lower, " error") || strings.Contains(lower, "[error]") {
			state.errors["messages"]++
		}
		if strings.Contains(lower, " warn") || strings.Contains(lower, "[warn]") {
			state.warnings["messages"]++
		}
	})
}
