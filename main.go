package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

const (
	namespace = "nessus"
)

// getDefaultLogPath returns OS-specific default log path for a given filename
func getDefaultLogPath(filename string) string {
	var baseDir string
	switch runtime.GOOS {
	case "linux":
		baseDir = "/opt/nessus/var/nessus/logs"
	case "darwin":
		baseDir = "/Library/Nessus/run/var/nessus/logs"
	case "windows":
		baseDir = `C:\ProgramData\Tenable\Nessus\nessus\logs`
	default:
		baseDir = "/opt/nessus/var/nessus/logs" // fallback to Linux
	}
	return filepath.Join(baseDir, filename)
}

func main() {
	kingpin.Version(version.Print("nessus_log_exporter"))

	// ── flag declarations with OS-detected defaults ─────────────
	listenAddr := kingpin.Flag("web.listen-address", "Address to listen on for metrics.").
		Default(":19835").Envar("LISTEN_ADDRESS").String()
	metricsPath := kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").
		Default("/metrics").Envar("TELEMETRY_PATH").String()

	messagesLog := kingpin.Flag("nessus.messages-log",
		"Path to nessusd.messages log file.").
		Default(getDefaultLogPath("nessusd.messages")).
		Envar("NESSUS_MESSAGES_LOG").String()
	backendLog := kingpin.Flag("nessus.backend-log",
		"Path to backend.log file.").
		Default(getDefaultLogPath("backend.log")).
		Envar("NESSUS_BACKEND_LOG").String()
	dumpLog := kingpin.Flag("nessus.dump-log",
		"Path to nessusd.dump file.").
		Default(getDefaultLogPath("nessusd.dump")).
		Envar("NESSUS_DUMP_LOG").String()
	cliLog := kingpin.Flag("nessus.cli-log",
		"Path to nessuscli.log file.").
		Default(getDefaultLogPath("nessuscli.log")).
		Envar("NESSUS_CLI_LOG").String()

	enableHostMetrics := kingpin.Flag("host.enable",
		"Enable host-level CPU/memory/disk/net metrics (gopsutil).").
		Default("true").Envar("HOST_METRICS_ENABLE").Bool()

	kingpin.Parse()

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	
	level.Info(logger).Log("msg", "Starting nessus_log_exporter", "version", version.Version)
	level.Info(logger).Log("msg", "Detected OS", "os", runtime.GOOS)
	level.Info(logger).Log(
		"msg", "Using log paths",
		"messages", *messagesLog,
		"backend", *backendLog,
		"dump", *dumpLog,
		"cli", *cliLog,
	)

	// ── collector ────────────────────────────────────────────────
	collector := NewNessusLogCollector(logger, LogPaths{
		Messages: *messagesLog,
		Backend:  *backendLog,
		Dump:     *dumpLog,
		CLI:      *cliLog,
	}, *enableHostMetrics)

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(collector)
	registry.MustRegister(version.NewCollector("nessus_log_exporter"))

	// ── HTTP ─────────────────────────────────────────────────────
	mux := http.NewServeMux()
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `<html><head><title>Nessus Log Exporter</title></head>
<body><h1>Nessus Log Exporter</h1>
<p><a href="%s">Metrics</a></p></body></html>`, *metricsPath)
	})

	level.Info(logger).Log("msg", "Listening", "addr", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, mux); err != nil {
		level.Error(logger).Log("msg", "HTTP server error", "err", err)
		os.Exit(1)
	}
}
