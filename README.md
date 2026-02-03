# Nessus Log Exporter

A production-ready Prometheus exporter that parses Nessus log files and exposes metrics for monitoring scan activity, errors, agent state, and host resources.

## Installation

### Quick Install (Linux)

Complete installation with systemd service:

```bash
# 1. Download and extract binary
wget https://github.com/k0wl0n/nessus_log_exporter/releases/download/v1.0.0/nessus_log_exporter-1.0.0-linux-amd64.tar.gz
tar xzf nessus_log_exporter-1.0.0-linux-amd64.tar.gz

# 2. Install binary
sudo mv nessus_log_exporter /usr/local/bin/
sudo chmod +x /usr/local/bin/nessus_log_exporter

# 3. Create systemd service
sudo tee /etc/systemd/system/nessus-exporter.service > /dev/null <<'EOF'
[Unit]
Description=Nessus Log Exporter
Documentation=https://github.com/k0wl0n/nessus_log_exporter
After=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/nessus_log_exporter \
    --nessus.messages-log=/opt/nessus/var/nessus/logs/nessusd.messages \
    --nessus.backend-log=/opt/nessus/var/nessus/logs/backend.log \
    --nessus.dump-log=/opt/nessus/var/nessus/logs/nessusd.dump \
    --nessus.cli-log=/opt/nessus/var/nessus/logs/nessuscli.log \
    --host.enable \
    --web.listen-address=:19835 \
    --web.telemetry-path=/metrics

Restart=always
RestartSec=10

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadOnlyPaths=/opt/nessus/var/nessus/logs

[Install]
WantedBy=multi-user.target
EOF

# 4. Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable nessus-exporter
sudo systemctl start nessus-exporter

# 5. Check status
sudo systemctl status nessus-exporter

# 6. View metrics
curl http://localhost:19835/metrics | grep nessus_
```

### Platform-Specific Downloads

Download the latest release for your platform from the [releases page](https://github.com/k0wl0n/nessus_log_exporter/releases):

**Linux AMD64**
```bash
wget https://github.com/k0wl0n/nessus_log_exporter/releases/download/v1.0.0/nessus_log_exporter-1.0.0-linux-amd64.tar.gz
tar xzf nessus_log_exporter-1.0.0-linux-amd64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/
sudo chmod +x /usr/local/bin/nessus_log_exporter
```

**Linux ARM64**
```bash
wget https://github.com/k0wl0n/nessus_log_exporter/releases/download/v1.0.0/nessus_log_exporter-1.0.0-linux-arm64.tar.gz
tar xzf nessus_log_exporter-1.0.0-linux-arm64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/
sudo chmod +x /usr/local/bin/nessus_log_exporter
```

**macOS AMD64 (Intel)**
```bash
wget https://github.com/k0wl0n/nessus_log_exporter/releases/download/v1.0.0/nessus_log_exporter-1.0.0-darwin-amd64.tar.gz
tar xzf nessus_log_exporter-1.0.0-darwin-amd64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/
sudo chmod +x /usr/local/bin/nessus_log_exporter
```

**macOS ARM64 (Apple Silicon)**
```bash
wget https://github.com/k0wl0n/nessus_log_exporter/releases/download/v1.0.0/nessus_log_exporter-1.0.0-darwin-arm64.tar.gz
tar xzf nessus_log_exporter-1.0.0-darwin-arm64.tar.gz
sudo mv nessus_log_exporter /usr/local/bin/
sudo chmod +x /usr/local/bin/nessus_log_exporter
```

**Windows AMD64**
```powershell
# Download from: https://github.com/k0wl0n/nessus_log_exporter/releases/download/v1.0.0/nessus_log_exporter-1.0.0-windows-amd64.zip
# Extract and add to PATH
```

### Alternative Installation Methods

**Install with Go**
```bash
go install github.com/k0wl0n/nessus_log_exporter@latest
```

**Build from Source**
```bash
git clone https://github.com/k0wl0n/nessus_log_exporter.git
cd nessus_log_exporter
go build -o nessus_log_exporter .
sudo mv nessus_log_exporter /usr/local/bin/
```

## Features

- **Zero-config deployment**: Auto-detects OS and uses appropriate default log paths
- **Cross-platform support**: Linux, macOS, and Windows
- **30+ Prometheus metrics**: Scan lifecycle, errors, warnings, agent state, host resources
- **Log rotation handling**: Automatic detection and recovery
- **No API required**: Works with Nessus Agent in air-gapped environments
- **Docker-ready**: Complete stack with Prometheus and Grafana

## Quick Start

After installation, the exporter will automatically start monitoring your Nessus logs:

```bash
# Check service status
sudo systemctl status nessus-exporter

# View logs
sudo journalctl -u nessus-exporter -f

# Test metrics endpoint
curl http://localhost:19835/metrics | grep nessus_

# Manual run (for testing)
nessus_log_exporter --help
```

### Configuration

The exporter uses OS-aware defaults but can be customized:

```bash
# Run with custom log paths
nessus_log_exporter \
  --nessus.messages-log=/custom/path/nessusd.messages \
  --nessus.backend-log=/custom/path/backend.log \
  --nessus.dump-log=/custom/path/nessusd.dump \
  --nessus.cli-log=/custom/path/nessuscli.log \
  --host.enable \
  --web.listen-address=:19835
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

### Example Metrics Output

```prometheus
# HELP nessus_agent_linked 1 if agent is currently linked to a manager, 0 otherwise.
# TYPE nessus_agent_linked gauge
nessus_agent_linked 1

# HELP nessus_backend_log_lines_total Total lines parsed from backend.log.
# TYPE nessus_backend_log_lines_total counter
nessus_backend_log_lines_total 1662

# HELP nessus_cli_commands_total Total nessuscli commands executed.
# TYPE nessus_cli_commands_total counter
nessus_cli_commands_total{command="fix"} 2
nessus_cli_commands_total{command="managed"} 1

# HELP nessus_errors_total Total error-level log messages.
# TYPE nessus_errors_total counter
nessus_errors_total{log_source="backend"} 2
nessus_errors_total{log_source="dump"} 1

# HELP nessus_host_cpu_percent Host CPU utilisation percent.
# TYPE nessus_host_cpu_percent gauge
nessus_host_cpu_percent 32.84

# HELP nessus_host_memory_percent Host memory utilisation percent.
# TYPE nessus_host_memory_percent gauge
nessus_host_memory_percent 5.61

# HELP nessus_host_disk_used_bytes Disk used bytes per mount.
# TYPE nessus_host_disk_used_bytes gauge
nessus_host_disk_used_bytes{mount="/"} 1.38e+10
nessus_host_disk_used_bytes{mount="/boot/efi"} 6.34e+06

# HELP nessus_host_net_recv_bytes_total Network bytes received per interface.
# TYPE nessus_host_net_recv_bytes_total counter
nessus_host_net_recv_bytes_total{interface="ens4"} 6.92e+09
nessus_host_net_recv_bytes_total{interface="lo"} 9.48e+08

# HELP nessus_info Nessus daemon version info.
# TYPE nessus_info gauge
nessus_info{build="20021",version="10.11.1"} 1

# HELP nessus_process_state 1 if nessusd is running, 0 otherwise.
# TYPE nessus_process_state gauge
nessus_process_state 1

# HELP nessus_messages_log_lines_total Total lines parsed from nessusd.messages.
# TYPE nessus_messages_log_lines_total counter
nessus_messages_log_lines_total 292
```

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
