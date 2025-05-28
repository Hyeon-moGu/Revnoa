package agent

import (
	"context"
	"revnoa/collector"
	"revnoa/config"
	"revnoa/sender"
	"revnoa/utils"
	"time"
)

func StartMetricsLoop(ctx context.Context, cfg *config.Config, agentID string) {
	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			utils.InfoLogger.Println("Stopped collecting metrics.")
			return

		case <-ticker.C:
			var cpu *collector.CPUStats
			var memory *collector.MemoryStats
			var disks []collector.DiskUsage
			var netStats *collector.NetStats
			var ports []collector.PortInfo
			var host *collector.HostInfo
			var dockerInfo []collector.DockerContainerInfo
			var redisInfo *collector.RedisMetrics

			if cfg.Collectors.CPU.Enabled {
				c, err := collector.CollectCpu()
				if err != nil {
					utils.ErrorLogger.Printf("Couldn't get CPU stats: %v", err)
				} else {
					cpu = c
				}
			}

			if cfg.Collectors.Mem.Enabled {
				m, err := collector.CollectMemory()
				if err != nil {
					utils.ErrorLogger.Printf("Couldn't get memory stats: %v", err)
				} else {
					memory = m
				}
			}

			if cfg.Collectors.Disk.Enabled {
				d, err := collector.CollectDisks()
				if err != nil {
					utils.ErrorLogger.Printf("Couldn't get disk usage: %v", err)
				} else {
					disks = d
				}
			}

			if cfg.Collectors.Net.Enabled {
				n, err := collector.CollectNetStats()
				if err != nil {
					utils.ErrorLogger.Printf("Couldn't get network info: %v", err)
				} else {
					netStats = n
				}
			}

			if cfg.Collectors.Ports.Enabled {
				p, err := collector.CollectOpenPorts()
				if err != nil {
					utils.ErrorLogger.Printf("Couldn't get open ports: %v", err)
				} else {
					ports = p
				}
			}

			if cfg.Collectors.Host.Enabled {
				h, err := collector.CollectHostInfo()
				if err != nil {
					utils.ErrorLogger.Printf("Couldn't get host info: %v", err)
				} else {
					host = h
				}
			}

			if cfg.Collectors.Docker.Enabled {
				info, err := collector.CollectDockerContainers()
				if err != nil {
					utils.WarnLogger.Printf("Docker check failed: %v", err)
				} else {
					dockerInfo = info
				}
			}

			if cfg.Collectors.Redis.Enabled {
				redisCollector := collector.NewRedisCollector(cfg.Collectors.Redis.Addr)
				info, err := redisCollector.Collect()
				if err != nil {
					utils.WarnLogger.Printf("Redis check failed: %v", err)
				} else {
					redisInfo = info
				}
			}

			payload := collector.FullMetrics{
				AgentID:   agentID,
				Cpu:       cpu,
				Memory:    memory,
				Disks:     disks,
				Net:       netStats,
				Ports:     ports,
				Host:      host,
				Docker:    dockerInfo,
				Redis:     redisInfo,
				Timestamp: time.Now().Unix(),
			}

			sender.SendMetricsLoop(cfg.API.Server, payload)
		}
	}
}
