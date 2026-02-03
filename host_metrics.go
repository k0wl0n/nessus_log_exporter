package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

func emitHostMetrics(c *NessusLogCollector, ch chan<- prometheus.Metric) {
	// ── CPU ────────────────────────────────────────────────────
	if pcts, err := cpu.Percent(0, false); err == nil && len(pcts) > 0 {
		ch <- prometheus.MustNewConstMetric(c.hostCPUPercent, prometheus.GaugeValue, pcts[0])
	}

	// ── Memory ─────────────────────────────────────────────────
	if v, err := mem.VirtualMemory(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.hostMemPercent, prometheus.GaugeValue, v.UsedPercent)
	}

	// ── Disk (all partitions) ──────────────────────────────────
	if parts, err := disk.Partitions(false); err == nil {
		for _, p := range parts {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.hostDiskUsedBytes, prometheus.GaugeValue, float64(usage.Used), p.Mountpoint)
			ch <- prometheus.MustNewConstMetric(c.hostDiskTotalBytes, prometheus.GaugeValue, float64(usage.Total), p.Mountpoint)
		}
	}

	// ── Network (all interfaces) ──────────────────────────────
	if counters, err := net.IOCounters(true); err == nil {
		for _, nic := range counters {
			ch <- prometheus.MustNewConstMetric(c.hostNetRecvBytes, prometheus.CounterValue, float64(nic.BytesRecv), nic.Name)
			ch <- prometheus.MustNewConstMetric(c.hostNetSendBytes, prometheus.CounterValue, float64(nic.BytesSent), nic.Name)
		}
	}

	// ── Uptime ─────────────────────────────────────────────────
	if up, err := host.Uptime(); err == nil {
		ch <- prometheus.MustNewConstMetric(c.hostUptime, prometheus.GaugeValue, float64(up))
	}
}
