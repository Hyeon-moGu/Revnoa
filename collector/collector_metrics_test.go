package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCollectCpu(t *testing.T) {
	cpuStats, err := CollectCpu()
	assert.NoError(t, err)
	assert.NotNil(t, cpuStats)
	assert.GreaterOrEqual(t, cpuStats.Cores, 1)
	assert.GreaterOrEqual(t, cpuStats.UsagePercent, 0.0)
}

func TestCollectMemory(t *testing.T) {
	memStats, err := CollectMemory()
	assert.NoError(t, err)
	assert.NotNil(t, memStats)
	assert.Greater(t, memStats.Total, uint64(0))
	assert.GreaterOrEqual(t, memStats.UsedPercent, 0.0)
	assert.LessOrEqual(t, memStats.UsedPercent, 100.0)
}

func TestCollectDisks(t *testing.T) {
	disks, err := CollectDisks()
	assert.NoError(t, err)
	assert.NotNil(t, disks)
	// allow empty slice
}

func TestCollectNetStats(t *testing.T) {
	netStats, err := CollectNetStats()
	assert.NoError(t, err)
	assert.NotNil(t, netStats)
	assert.GreaterOrEqual(t, netStats.BytesSent, uint64(0))
}

func TestCollectHostInfo(t *testing.T) {
	host, err := CollectHostInfo()
	assert.NoError(t, err)
	assert.NotNil(t, host)
	assert.NotEmpty(t, host.Hostname)
	assert.NotEmpty(t, host.OS)
}

func TestCollectOpenPorts(t *testing.T) {
	ports, err := CollectOpenPorts()
	// allow errors if command not available
	if err == nil {
		assert.NotNil(t, ports)
	}
}
