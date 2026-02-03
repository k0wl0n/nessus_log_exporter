# Nessus Log Exporter

A production-ready Prometheus exporter that parses Nessus log files and exposes metrics for monitoring scan activity, errors, agent state, and host resources.

## Installation

### Option 1: Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/k0wl0n/nessus_log_exporter/releases):

```bash
# Linux AMD64
wget https://github.com/k0wl0n/nessus_log_exporter/releases/latest/download/nessus_log_exporter-linux-amd64.tar.gz
tar xzf nessus_log_exporter-linux-amd64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/

# Linux ARM64
wget https://github.com/k0wl0n/nessus_log_exporter/releases/latest/download/nessus_log_exporter-linux-arm64.tar.gz
tar xzf nessus_log_exporter-linux-arm64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/

# macOS AMD64
wget https://github.com/k0wl0n/nessus_log_exporter/releases/latest/download/nessus_log_exporter-darwin-amd64.tar.gz
tar xzf nessus_log_exporter-darwin-amd64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/

# macOS ARM64 (Apple Silicon)
wget https://github.com/k0wl0n/nessus_log_exporter/releases/latest/download/nessus_log_exporter-darwin-arm64.tar.gz
tar xzf nessus_log_exporter-darwin-arm64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/

# Windows AMD64
# Download nessus_log_exporter-windows-amd64.zip from releases page
```

### Option 2: Install with Go

```bash
go install github.com/k0wl0n/nessus_log_exporter@latest
```

### Option 3: Build from Source

```bash
git clone https://github.com/k0wl0n/nessus_log_exporter.git
cd nessus_log_exporter
go build -o nessus_log_exporter .
```

## Features

- **Zero-config deployment**: Auto-detects OS and uses appropriate default log paths
- **Cross-platform support**: Linux, macOS, and Windows
- **30+ Prometheus metrics**: Scan lifecycle, errors, warnings, agent state, host resources
- **Log rotation handling**: Automatic detection and recovery
- **No API required**: Works with Nessus Agent in air-gapped environments
- **Docker-ready**: Complete stack with Prometheus and Grafana

## Quick Start

### Option 1: With Nessus Scanner in Docker (Recommended)

```bash
# 1. Start complete stack (Nessus + Exporter + Prometheus + Grafana)
docker compose up -d

# 2. Configure Nessus
open https://localhost:8834  # Nessus Web UI (see NESSUS_SETUP.md)

# 3. Access monitoring dashboards
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
open http://localhost:19835/metrics  # Exporter metrics
```

**📖 See [NESSUS_SETUP.md](NESSUS_SETUP.md) for detailed Nessus Docker configuration.**

### Option 2: With Existing Nessus Installation

```bash
# 1. Clone and build
cd nessus_exporter
go mod tidy

# 2. Edit docker-compose.yml to mount your existing Nessus logs
# 3. Start the monitoring stack
docker compose up -d --build

# 4. Access dashboards
open http://localhost:3000  # Grafana (admin/admin)
open http://localhost:9090  # Prometheus
open http://localhost:19835/metrics  # Exporter metrics
```

## OS Detection

The exporter automatically detects your operating system and uses the appropriate default log paths:

| OS | Default Log Path |
|---|---|
| **Linux** | `/opt/nessus/var/nessus/logs/` |
| **macOS** | `/Library/Nessus/run/var/nessus/logs/` |
| **Windows** | `C:\ProgramData\Tenable\Nessus\nessus\logs\` |

## Configuration Priority

The exporter uses a three-tier configuration system:

1. **Environment Variables** (highest priority)
2. **CLI Flags** (medium priority)
3. **OS-Detected Defaults** (lowest priority, automatic)

### Environment Variables

```bash
export NESSUS_MESSAGES_LOG=/custom/path/nessusd.messages
export NESSUS_BACKEND_LOG=/custom/path/backend.log
export NESSUS_DUMP_LOG=/custom/path/nessusd.dump
export NESSUS_CLI_LOG=/custom/path/nessuscli.log
export HOST_METRICS_ENABLE=true
```

### CLI Flags

```bash
./nessus_log_exporter \
  --nessus.messages-log=/custom/path/nessusd.messages \
  --nessus.backend-log=/custom/path/backend.log \
  --nessus.dump-log=/custom/path/nessusd.dump \
  --nessus.cli-log=/custom/path/nessuscli.log \
  --host.enable=true \
  --web.listen-address=:19835 \
  --web.telemetry-path=/metrics
```

## Docker Deployment

### Volume Mounting

Edit `docker-compose.yml` to uncomment the volume mount for your OS:

```yaml
volumes:
  # Linux (default)
  - /opt/nessus/var/nessus/logs:/opt/nessus/var/nessus/logs:ro
  
  # macOS
  # - /Library/Nessus/run/var/nessus/logs:/Library/Nessus/run/var/nessus/logs:ro
  
  # Windows
  # - C:\ProgramData\Tenable\Nessus\nessus\logs:C:\ProgramData\Tenable\Nessus\nessus\logs:ro
```

### Environment Variable Override

```yaml
environment:
  - NESSUS_MESSAGES_LOG=/custom/path/nessusd.messages
  - HOST_METRICS_ENABLE=true
```

## Metrics

### Scan Lifecycle (6 metrics)
- `nessus_scan_started_total{scan_name, user}` - Total scans started
- `nessus_scan_completed_total{scan_name}` - Total scans completed
- `nessus_scan_duration_seconds{scan_name, host}` - Last scan duration per host
- `nessus_hosts_tested_total` - Total hosts tested
- `nessus_hosts_up_total` - Total hosts reported UP
- `nessus_plugins_launched_total{scan_name}` - Total NASL plugins launched

### Errors & Warnings (2 metrics)
- `nessus_errors_total{log_source}` - Error count by log source
- `nessus_warnings_total{log_source}` - Warning count by log source

### Raw Line Counters (4 metrics)
- `nessus_messages_log_lines_total` - Lines parsed from nessusd.messages
- `nessus_backend_log_lines_total` - Lines parsed from backend.log
- `nessus_dump_log_lines_total` - Lines parsed from nessusd.dump
- `nessus_cli_log_lines_total` - Lines parsed from nessuscli.log

### CLI Commands (1 metric)
- `nessus_cli_commands_total{command}` - nessuscli commands executed

### Agent State (3 metrics)
- `nessus_agent_linked` - 1 if agent is linked, 0 otherwise
- `nessus_process_state` - 1 if nessusd is running, 0 otherwise
- `nessus_info{version, build}` - Nessus version info

### Host Metrics (7 metrics)
- `nessus_host_cpu_percent` - Host CPU utilization
- `nessus_host_memory_percent` - Host memory utilization
- `nessus_host_disk_used_bytes{mount}` - Disk used bytes per mount
- `nessus_host_disk_total_bytes{mount}` - Disk total bytes per mount
- `nessus_host_net_recv_bytes_total{interface}` - Network bytes received
- `nessus_host_net_send_bytes_total{interface}` - Network bytes sent
- `nessus_host_uptime_seconds` - Host uptime

## Nessus Configuration

### For Plugin Launch Metrics

Enable "Log scan details" in scan profile → Advanced settings. Without this, `nessus_plugins_launched_total` will remain at 0.

### For NASL Dump Metrics

Set NASL log level to trace or full:

```bash
nessuscli fix --set nasl_log_type=trace
```

**Warning**: `trace` and `full` generate very large log files during scans. Monitor disk usage.

### Optional: Millisecond Timestamps

```bash
nessuscli fix --set logfile_msec=yes
```

## Prometheus Alerts

The exporter includes 8 pre-configured alert rules:

- **Exporter Issues**: https://github.com/k0wl0n/nessus_log_exporter/issues
- **NessusProcessNotRunning** - nessusd not detected
- **NessusAgentUnlinked** - Agent not linked to manager
- **NessusHighErrorRate** - Error rate > 2/min
- **NessusScanStartedButNotCompleted** - Scan running > 30 min
- **NessusHostHighCPU** - CPU > 85%
- **NessusHostHighMemory** - Memory > 85%
- **NessusHostDiskLow** - Disk > 90% full

## Grafana Dashboard

The included dashboard provides:

- **Status panels**: Scan counts, process state, agent link status
- **Time series**: Scan rates, error rates
- **Host metrics**: CPU, memory, disk usage

Dashboard auto-provisions on first Grafana startup.

## Architecture

```
┌─────────────────┐
│ Nessus Scanner  │
│  Log Files      │
└────────┬────────┘
         │ (read-only)
         ▼
┌─────────────────┐
│ Log Exporter    │
│  - tail.go      │
│  - parsers      │
│  - collector    │
└────────┬────────┘
         │ :19835/metrics
         ▼
┌─────────────────┐
│  Prometheus     │
│  - scrape 15s   │
│  - alerts       │
└────────┬────────┘
         │ :9090
         ▼
┌─────────────────┐
│    Grafana      │
│  - dashboards   │
└─────────────────┘
   :3000
```

## Development

### Build

```bash
go build -o nessus_log_exporter .
```

### Run Locally

```bash
./nessus_log_exporter \
  --nessus.messages-log=./nessusd.messages \
  --nessus.backend-log=./backend.log \
  --nessus.dump-log=./nessusd.dump \
  --nessus.cli-log=./nessuscli.log
```

### Test with Sample Logs

Place sample log files in the project directory and run the exporter pointing to them.

## Known Limitations

1. **In-memory state only**: Restart loses historical counts (Prometheus handles this with `rate()`)
2. **No historical replay**: Only tails new log lines since exporter start
3. **Scan name extraction**: Requires `[JOB_NAME=...]` format in nessusd.messages
4. **Plugin launch counting**: Requires "Log scan details" enabled in scan profile
5. **NASL dump metrics**: Requires `nasl_log_type=trace` or `full` (generates large files)
6. **Duration is last-seen**: `scan_duration_seconds` only keeps most recent value per scan+host
7. **Log rotation detection**: If file shrinks, assumes rotation and resets to offset 0

## Troubleshooting

### No metrics appearing

1. Check exporter logs: `docker logs nessus_log_exporter`
2. Verify log paths are correct and readable
3. Ensure Nessus is generating log activity

### Plugin metrics at zero

Enable "Log scan details" in Nessus scan profile → Advanced settings.

### NASL dump metrics sparse

Set `nasl_log_type=trace` via `nessuscli fix --set nasl_log_type=trace`.

### Permission denied errors

Ensure the exporter has read access to Nessus log files. On Linux, you may need to run the container with appropriate user/group permissions.

## License

MIT

## Contributing

Contributions welcome! Please open an issue or pull request.
