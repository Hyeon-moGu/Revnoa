package collector

import (
	"context"
	"strings"
	"time"

	"revnoa/utils"

	"github.com/go-redis/redis/v8"
)

type RedisCollector struct {
	client *redis.Client
	ctx    context.Context
	addr   string
}

type RedisMetrics struct {
	Status           string `json:"status"`
	ConnectedClients string `json:"connected_clients,omitempty"`
	UsedMemory       string `json:"used_memory,omitempty"`
	TotalConnections string `json:"total_connections_received,omitempty"`
	UptimeInSeconds  string `json:"uptime_in_seconds,omitempty"`
	Role             string `json:"role,omitempty"`
	RedisVersion     string `json:"redis_version,omitempty"`
}

func NewRedisCollector(addr string) *RedisCollector {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisCollector{client: client, ctx: ctx, addr: addr}
}

func (rc *RedisCollector) Collect() (*RedisMetrics, error) {
	start := time.Now()

	info, err := rc.client.Info(rc.ctx, "all").Result()
	if err != nil {
		utils.WarnLogger.Printf("Can't connect to Redis at %s: %v", rc.addr, err)
		return &RedisMetrics{Status: "unreachable"}, nil
	}

	metrics := &RedisMetrics{Status: "ok"}

	for _, line := range strings.Split(info, "\n") {
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "connected_clients":
			metrics.ConnectedClients = value
		case "used_memory":
			metrics.UsedMemory = value
		case "total_connections_received":
			metrics.TotalConnections = value
		case "uptime_in_seconds":
			metrics.UptimeInSeconds = value
		case "role":
			metrics.Role = value
		case "redis_version":
			metrics.RedisVersion = value
		}
	}

	utils.InfoLogger.Printf("Got Redis metrics in %v", time.Since(start))
	return metrics, nil
}
