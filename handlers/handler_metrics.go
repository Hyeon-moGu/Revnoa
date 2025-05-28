package handlers

import (
	"encoding/json"
	"net/http"
	"revnoa/collector"
	"revnoa/config"
	"revnoa/utils"
	"time"
)

func GetMetricsHandler(agentID string, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthorized(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		utils.InfoLogger.Println("Received /metrics request")

		cpu := tryCollectCPU(cfg)
		mem := tryCollectMemory(cfg)
		disks := tryCollectDisks(cfg)
		netStats := tryCollectNet(cfg)
		ports := tryCollectPorts(cfg)
		host := tryCollectHost(cfg)
		docker := tryCollectDocker(cfg)
		redis := tryCollectRedis(cfg)

		payload := collector.FullMetrics{
			AgentID:   agentID,
			Cpu:       cpu,
			Memory:    mem,
			Disks:     disks,
			Net:       netStats,
			Ports:     ports,
			Host:      host,
			Docker:    docker,
			Redis:     redis,
			Timestamp: time.Now().Unix(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(payload); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			utils.ErrorLogger.Printf("JSON encode error: %v", err)
		}
	}
}

func tryCollectCPU(cfg *config.Config) *collector.CPUStats {
	if !cfg.Collectors.CPU.Enabled {
		return nil
	}
	c, err := collector.CollectCpu()
	if err != nil {
		utils.ErrorLogger.Println("CPU collection error:", err)
		return nil
	}
	return c
}

func tryCollectMemory(cfg *config.Config) *collector.MemoryStats {
	if !cfg.Collectors.Mem.Enabled {
		return nil
	}
	m, err := collector.CollectMemory()
	if err != nil {
		utils.ErrorLogger.Println("Memory collection error:", err)
		return nil
	}
	return m
}

func tryCollectDisks(cfg *config.Config) []collector.DiskUsage {
	if !cfg.Collectors.Disk.Enabled {
		return nil
	}
	d, err := collector.CollectDisks()
	if err != nil {
		utils.ErrorLogger.Println("Disk collection error:", err)
		return nil
	}
	return d
}

func tryCollectNet(cfg *config.Config) *collector.NetStats {
	if !cfg.Collectors.Net.Enabled {
		return nil
	}
	n, err := collector.CollectNetStats()
	if err != nil {
		utils.ErrorLogger.Println("Network collection error:", err)
		return nil
	}
	return n
}

func tryCollectPorts(cfg *config.Config) []collector.PortInfo {
	if !cfg.Collectors.Ports.Enabled {
		return nil
	}
	p, err := collector.CollectOpenPorts()
	if err != nil {
		utils.ErrorLogger.Println("Port collection error:", err)
		return nil
	}
	return p
}

func tryCollectHost(cfg *config.Config) *collector.HostInfo {
	if !cfg.Collectors.Host.Enabled {
		return nil
	}
	h, err := collector.CollectHostInfo()
	if err != nil {
		utils.ErrorLogger.Println("Host info collection error:", err)
		return nil
	}
	return h
}

func tryCollectDocker(cfg *config.Config) []collector.DockerContainerInfo {
	if !cfg.Collectors.Docker.Enabled {
		return nil
	}
	containers, _ := collector.CollectDockerContainers()
	return containers
}

func tryCollectRedis(cfg *config.Config) *collector.RedisMetrics {
	if !cfg.Collectors.Redis.Enabled {
		return nil
	}
	redisCollector := collector.NewRedisCollector(cfg.Collectors.Redis.Addr)
	info, err := redisCollector.Collect()
	if err != nil {
		utils.WarnLogger.Printf("Redis collection error: %v", err)
		return nil
	}
	return info
}
