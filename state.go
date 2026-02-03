package main

// exporterState lives for the lifetime of the process.
// Counters accumulate; gauges are overwritten each Collect().
type exporterState struct {
	// ── file-tail cursors (byte offset per path) ────────────────
	offsets map[string]int64

	// ── scan lifecycle ───────────────────────────────────────────
	scanStarted      map[scanUserKey]int     // {scan_name, user} -> count
	scanCompleted    map[string]int          // scan_name -> count
	lastScanDuration map[scanHostKey]float64 // {scan_name, host} -> seconds
	hostsTested      int
	hostsUp          int

	// ── plugins ──────────────────────────────────────────────────
	pluginsLaunched map[string]int // scan_name -> count

	// ── errors / warnings per log source ─────────────────────────
	errors   map[string]int // log_source (messages|backend|dump|cli)
	warnings map[string]int

	// ── raw line counts ──────────────────────────────────────────
	messagesLines int
	backendLines  int
	dumpLines     int
	cliLines      int

	// ── nessuscli commands ───────────────────────────────────────
	cliCommands map[string]int // command name -> count

	// ── daemon / agent state (gauges, overwritten) ──────────────
	isRunning     bool
	isAgentLinked bool
	daemonVersion string
	daemonBuild   string
}

type scanUserKey struct {
	scanName string
	user     string
}

type scanHostKey struct {
	scanName string
	host     string
}

func newExporterState() *exporterState {
	return &exporterState{
		offsets:          make(map[string]int64),
		scanStarted:      make(map[scanUserKey]int),
		scanCompleted:    make(map[string]int),
		lastScanDuration: make(map[scanHostKey]float64),
		pluginsLaunched:  make(map[string]int),
		errors:           make(map[string]int),
		warnings:         make(map[string]int),
		cliCommands:      make(map[string]int),
	}
}
