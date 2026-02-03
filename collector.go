package main

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

// LogPaths groups every log path the exporter reads.
type LogPaths struct {
	Messages string // nessusd.messages
	Backend  string // backend.log
	Dump     string // nessusd.dump
	CLI      string // nessuscli.log
}

// NessusLogCollector implements prometheus.Collector.
// Every Collect() call re-tails the logs so metrics stay fresh.
type NessusLogCollector struct {
	logger          log.Logger
	paths           LogPaths
	enableHostStats bool

	// ── Nessus scan lifecycle ──────────────────────────────────
	scanStartedTotal   *prometheus.Desc // counter
	scanCompletedTotal *prometheus.Desc // counter
	scanDurationSecs   *prometheus.Desc // gauge  – last seen duration per scan_name
	hostsTestedTotal   *prometheus.Desc // counter
	hostsUpTotal       *prometheus.Desc // counter

	// ── Plugin activity ────────────────────────────────────────
	pluginsLaunchedTotal *prometheus.Desc // counter

	// ── Error / warning counters (per log source) ─────────────
	errorsTotal   *prometheus.Desc // counter {log_source}
	warningsTotal *prometheus.Desc // counter {log_source}

	// ── Raw line counters ─────────────────────────────────────
	backendLinesTotal  *prometheus.Desc
	messagesLinesTotal *prometheus.Desc
	dumpLinesTotal     *prometheus.Desc
	cliLinesTotal      *prometheus.Desc

	// ── nessuscli audit ───────────────────────────────────────
	cliCommandsTotal *prometheus.Desc // counter {command}

	// ── Agent state ────────────────────────────────────────────
	agentLinked   *prometheus.Desc // gauge 0|1
	processState  *prometheus.Desc // gauge 0|1
	nessusVersion *prometheus.Desc // gauge {version, build}

	// ── Host / OS utilisation ─────────────────────────────────
	hostCPUPercent     *prometheus.Desc
	hostMemPercent     *prometheus.Desc
	hostDiskUsedBytes  *prometheus.Desc // {mount}
	hostDiskTotalBytes *prometheus.Desc
	hostNetRecvBytes   *prometheus.Desc // {interface}
	hostNetSendBytes   *prometheus.Desc
	hostUptime         *prometheus.Desc

	// ── internal state (persists between scrapes for counters) ─
	state *exporterState
}

func NewNessusLogCollector(logger log.Logger, paths LogPaths, enableHost bool) *NessusLogCollector {
	return &NessusLogCollector{
		logger:          logger,
		paths:           paths,
		enableHostStats: enableHost,
		state:           newExporterState(),

		// scan
		scanStartedTotal: prometheus.NewDesc(
			qn("scan_started_total"), "Total number of scans started.", []string{"scan_name", "user"}, nil),
		scanCompletedTotal: prometheus.NewDesc(
			qn("scan_completed_total"), "Total number of scans completed.", []string{"scan_name"}, nil),
		scanDurationSecs: prometheus.NewDesc(
			qn("scan_duration_seconds"), "Duration in seconds of the last completed scan per host.", []string{"scan_name", "host"}, nil),
		hostsTestedTotal: prometheus.NewDesc(
			qn("hosts_tested_total"), "Total hosts tested across all scans.", nil, nil),
		hostsUpTotal: prometheus.NewDesc(
			qn("hosts_up_total"), "Total hosts reported UP.", nil, nil),

		// plugins
		pluginsLaunchedTotal: prometheus.NewDesc(
			qn("plugins_launched_total"), "Total NASL plugins launched.", []string{"scan_name"}, nil),

		// errors
		errorsTotal: prometheus.NewDesc(
			qn("errors_total"), "Total error-level log messages.", []string{"log_source"}, nil),
		warningsTotal: prometheus.NewDesc(
			qn("warnings_total"), "Total warning-level log messages.", []string{"log_source"}, nil),

		// raw lines
		backendLinesTotal: prometheus.NewDesc(
			qn("backend_log_lines_total"), "Total lines parsed from backend.log.", nil, nil),
		messagesLinesTotal: prometheus.NewDesc(
			qn("messages_log_lines_total"), "Total lines parsed from nessusd.messages.", nil, nil),
		dumpLinesTotal: prometheus.NewDesc(
			qn("dump_log_lines_total"), "Total lines parsed from nessusd.dump.", nil, nil),
		cliLinesTotal: prometheus.NewDesc(
			qn("cli_log_lines_total"), "Total lines parsed from nessuscli.log.", nil, nil),

		// cli commands
		cliCommandsTotal: prometheus.NewDesc(
			qn("cli_commands_total"), "Total nessuscli commands executed.", []string{"command"}, nil),

		// agent / daemon
		agentLinked: prometheus.NewDesc(
			qn("agent_linked"), "1 if agent is currently linked to a manager, 0 otherwise.", nil, nil),
		processState: prometheus.NewDesc(
			qn("process_state"), "1 if nessusd is running, 0 otherwise.", nil, nil),
		nessusVersion: prometheus.NewDesc(
			qn("info"), "Nessus daemon version info.", []string{"version", "build"}, nil),

		// host
		hostCPUPercent: prometheus.NewDesc(
			qn("host_cpu_percent"), "Host CPU utilisation percent.", nil, nil),
		hostMemPercent: prometheus.NewDesc(
			qn("host_memory_percent"), "Host memory utilisation percent.", nil, nil),
		hostDiskUsedBytes: prometheus.NewDesc(
			qn("host_disk_used_bytes"), "Disk used bytes per mount.", []string{"mount"}, nil),
		hostDiskTotalBytes: prometheus.NewDesc(
			qn("host_disk_total_bytes"), "Disk total bytes per mount.", []string{"mount"}, nil),
		hostNetRecvBytes: prometheus.NewDesc(
			qn("host_net_recv_bytes_total"), "Network bytes received per interface.", []string{"interface"}, nil),
		hostNetSendBytes: prometheus.NewDesc(
			qn("host_net_send_bytes_total"), "Network bytes sent per interface.", []string{"interface"}, nil),
		hostUptime: prometheus.NewDesc(
			qn("host_uptime_seconds"), "Host uptime in seconds.", nil, nil),
	}
}

// qn qualifies a metric name under the shared namespace.
func qn(name string) string { return namespace + "_" + name }

// ── prometheus.Collector interface ─────────────────────────────────

func (c *NessusLogCollector) Describe(ch chan<- *prometheus.Desc) {
	descs := []*prometheus.Desc{
		c.scanStartedTotal, c.scanCompletedTotal, c.scanDurationSecs,
		c.hostsTestedTotal, c.hostsUpTotal,
		c.pluginsLaunchedTotal,
		c.errorsTotal, c.warningsTotal,
		c.backendLinesTotal, c.messagesLinesTotal, c.dumpLinesTotal, c.cliLinesTotal,
		c.cliCommandsTotal,
		c.agentLinked, c.processState, c.nessusVersion,
		c.hostCPUPercent, c.hostMemPercent,
		c.hostDiskUsedBytes, c.hostDiskTotalBytes,
		c.hostNetRecvBytes, c.hostNetSendBytes,
		c.hostUptime,
	}
	for _, d := range descs {
		ch <- d
	}
}

func (c *NessusLogCollector) Collect(ch chan<- prometheus.Metric) {
	// ── 1. parse all logs → update state ────────────────────────
	if err := parseMessagesLog(c.paths.Messages, c.state, c.logger); err != nil {
		level.Warn(c.logger).Log("msg", "failed to parse nessusd.messages", "err", err)
	}
	if err := parseBackendLog(c.paths.Backend, c.state, c.logger); err != nil {
		level.Warn(c.logger).Log("msg", "failed to parse backend.log", "err", err)
	}
	if err := parseDumpLog(c.paths.Dump, c.state, c.logger); err != nil {
		level.Warn(c.logger).Log("msg", "failed to parse nessusd.dump", "err", err)
	}
	if err := parseCLILog(c.paths.CLI, c.state, c.logger); err != nil {
		level.Warn(c.logger).Log("msg", "failed to parse nessuscli.log", "err", err)
	}

	// ── 2. emit scan metrics ─────────────────────────────────────
	for key, val := range c.state.scanStarted {
		ch <- prometheus.MustNewConstMetric(c.scanStartedTotal, prometheus.CounterValue, float64(val), key.scanName, key.user)
	}
	for scanName, val := range c.state.scanCompleted {
		ch <- prometheus.MustNewConstMetric(c.scanCompletedTotal, prometheus.CounterValue, float64(val), scanName)
	}
	for key, dur := range c.state.lastScanDuration {
		ch <- prometheus.MustNewConstMetric(c.scanDurationSecs, prometheus.GaugeValue, dur, key.scanName, key.host)
	}
	ch <- prometheus.MustNewConstMetric(c.hostsTestedTotal, prometheus.CounterValue, float64(c.state.hostsTested))
	ch <- prometheus.MustNewConstMetric(c.hostsUpTotal, prometheus.CounterValue, float64(c.state.hostsUp))

	// ── 3. plugin counter ────────────────────────────────────────
	for scanName, cnt := range c.state.pluginsLaunched {
		ch <- prometheus.MustNewConstMetric(c.pluginsLaunchedTotal, prometheus.CounterValue, float64(cnt), scanName)
	}

	// ── 4. error / warning ───────────────────────────────────────
	for src, cnt := range c.state.errors {
		ch <- prometheus.MustNewConstMetric(c.errorsTotal, prometheus.CounterValue, float64(cnt), src)
	}
	for src, cnt := range c.state.warnings {
		ch <- prometheus.MustNewConstMetric(c.warningsTotal, prometheus.CounterValue, float64(cnt), src)
	}

	// ── 5. raw line counters ─────────────────────────────────────
	ch <- prometheus.MustNewConstMetric(c.backendLinesTotal, prometheus.CounterValue, float64(c.state.backendLines))
	ch <- prometheus.MustNewConstMetric(c.messagesLinesTotal, prometheus.CounterValue, float64(c.state.messagesLines))
	ch <- prometheus.MustNewConstMetric(c.dumpLinesTotal, prometheus.CounterValue, float64(c.state.dumpLines))
	ch <- prometheus.MustNewConstMetric(c.cliLinesTotal, prometheus.CounterValue, float64(c.state.cliLines))

	// ── 6. CLI commands ──────────────────────────────────────────
	for cmd, cnt := range c.state.cliCommands {
		ch <- prometheus.MustNewConstMetric(c.cliCommandsTotal, prometheus.CounterValue, float64(cnt), cmd)
	}

	// ── 7. agent / daemon state ──────────────────────────────────
	ch <- prometheus.MustNewConstMetric(c.agentLinked, prometheus.GaugeValue, boolToFloat(c.state.isAgentLinked))
	ch <- prometheus.MustNewConstMetric(c.processState, prometheus.GaugeValue, boolToFloat(c.state.isRunning))
	if c.state.daemonVersion != "" {
		ch <- prometheus.MustNewConstMetric(c.nessusVersion, prometheus.GaugeValue, 1, c.state.daemonVersion, c.state.daemonBuild)
	}

	// ── 8. host / OS metrics ─────────────────────────────────────
	if c.enableHostStats {
		emitHostMetrics(c, ch)
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
