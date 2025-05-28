package collector

import (
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

type FullMetrics struct {
	AgentID   string                `json:"agent_id"`
	Cpu       *CPUStats             `json:"cpu,omitempty"`
	Memory    *MemoryStats          `json:"memory,omitempty"`
	Disks     []DiskUsage           `json:"disks,omitempty"`
	Net       *NetStats             `json:"net,omitempty"`
	Ports     []PortInfo            `json:"ports,omitempty"`
	Host      *HostInfo             `json:"host,omitempty"`
	Timestamp int64                 `json:"timestamp"`
	Docker    []DockerContainerInfo `json:"docker,omitempty"`
	Redis     *RedisMetrics         `json:"redis,omitempty"`
}

type CPUStats struct {
	TimeUser   float64 `json:"time_user_seconds"`
	TimeSystem float64 `json:"time_system_seconds"`
	TimeIdle   float64 `json:"time_idle_seconds"`
	TimeIOWait float64 `json:"time_iowait_seconds,omitempty"`

	UsagePercent float64 `json:"usage_percent"`
	Cores        int     `json:"cores"`

	Load1  float64 `json:"load_1min"`
	Load5  float64 `json:"load_5min"`
	Load15 float64 `json:"load_15min"`
}

type MemoryStats struct {
	Total           uint64  `json:"total"`
	Used            uint64  `json:"used"`
	Free            uint64  `json:"free"`
	UsedPercent     float64 `json:"used_percent"`
	Cached          uint64  `json:"cached,omitempty"`
	Buffers         uint64  `json:"buffers,omitempty"`
	SwapTotal       uint64  `json:"swap_total"`
	SwapUsed        uint64  `json:"swap_used"`
	SwapUsedPercent float64 `json:"swap_used_percent"`
}

type DiskUsage struct {
	MountPoint string  `json:"mount_point"`
	Total      uint64  `json:"total"`
	Used       uint64  `json:"used"`
	UsedPerc   float64 `json:"used_perc"`
}

type NetStats struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

type PortInfo struct {
	Port string `json:"port"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
	Uptime   uint64 `json:"uptime"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
}

func CollectCpu() (*CPUStats, error) {
	cpuTimes, err := cpu.Times(false)
	if err != nil || len(cpuTimes) == 0 {
		return nil, err
	}
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil || len(cpuPercent) == 0 {
		return nil, err
	}
	cores, _ := cpu.Counts(true)
	loadAvg, _ := load.Avg()

	return &CPUStats{
		TimeUser:     cpuTimes[0].User,
		TimeSystem:   cpuTimes[0].System,
		TimeIdle:     cpuTimes[0].Idle,
		TimeIOWait:   cpuTimes[0].Iowait,
		UsagePercent: cpuPercent[0],
		Cores:        cores,
		Load1:        loadAvg.Load1,
		Load5:        loadAvg.Load5,
		Load15:       loadAvg.Load15,
	}, nil
}

func CollectMemory() (*MemoryStats, error) {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	swap, err := mem.SwapMemory()
	if err != nil {
		return nil, err
	}

	return &MemoryStats{
		Total:           vmem.Total,
		Used:            vmem.Used,
		Free:            vmem.Available,
		UsedPercent:     vmem.UsedPercent,
		SwapTotal:       swap.Total,
		SwapUsed:        swap.Used,
		SwapUsedPercent: swap.UsedPercent,
	}, nil
}

func CollectDisks() ([]DiskUsage, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}

	var result []DiskUsage
	for _, p := range partitions {
		if p.Fstype == "" || strings.HasPrefix(p.Fstype, "tmpfs") || strings.HasPrefix(p.Fstype, "dev") || strings.Contains(p.Device, "loop") {
			continue
		}

		usage, err := disk.Usage(p.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		result = append(result, DiskUsage{
			MountPoint: p.Mountpoint,
			Total:      usage.Total,
			Used:       usage.Used,
			UsedPerc:   usage.UsedPercent,
		})
	}
	return FilterUniqueDisks(result), nil
}

func CollectNetStats() (*NetStats, error) {
	counters, err := net.IOCounters(false)
	if err != nil || len(counters) == 0 {
		return nil, err
	}
	return &NetStats{
		BytesSent:   counters[0].BytesSent,
		BytesRecv:   counters[0].BytesRecv,
		PacketsSent: counters[0].PacketsSent,
		PacketsRecv: counters[0].PacketsRecv,
	}, nil
}

func CollectOpenPorts() ([]PortInfo, error) {
	osType := runtime.GOOS

	var output []byte
	var err error

	switch osType {
	case "windows":
		output, err = exec.Command("cmd", "/C", "netstat -an | findstr LISTENING").CombinedOutput()

	case "darwin", "linux":
		if isCommandAvailable("lsof") {
			output, err = exec.Command("lsof", "-i", "-nP", "-sTCP:LISTEN").CombinedOutput()
		} else if isCommandAvailable("ss") {
			output, err = exec.Command("ss", "-tuln").CombinedOutput()
		} else {
			return nil, fmt.Errorf("neither lsof nor ss is available")
		}

	default:
		return nil, fmt.Errorf("unsupported OS: %s", osType)
	}

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	portSet := make(map[string]struct{})

	var portRegex *regexp.Regexp
	if osType == "windows" {
		portRegex = regexp.MustCompile(`:(\d+)$`)
	} else if strings.Contains(string(output), "LISTEN") {
		portRegex = regexp.MustCompile(`:(\d+)\s+\(LISTEN\)`)
	} else {
		portRegex = regexp.MustCompile(`[:](\d+)\s`)
	}

	for _, line := range lines {
		matches := portRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			port := matches[1]
			if port != "0" {
				portSet[port] = struct{}{}
			}
		}
	}

	var ports []PortInfo
	for p := range portSet {
		ports = append(ports, PortInfo{Port: p})
	}

	return ports, nil
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func CollectHostInfo() (*HostInfo, error) {
	info, err := host.Info()
	if err != nil {
		return nil, err
	}

	return &HostInfo{
		Hostname: info.Hostname,
		Uptime:   info.Uptime,
		OS:       info.OS,
		Platform: info.Platform,
	}, nil
}

func FilterUniqueDisks(disks []DiskUsage) []DiskUsage {
	unique := make([]DiskUsage, 0)
	seen := make(map[string]bool)

	for _, d := range disks {
		// Unique key by Total and Used
		key := fmt.Sprintf("%d:%d", d.Total, d.Used)

		if !seen[key] {
			unique = append(unique, d)
			seen[key] = true
		}
	}

	return unique
}
