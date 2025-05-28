# Revnoa

**Revnoa** is a lightweight system monitoring agent written in Go.  
It periodically collects key system metrics — including CPU, memory, disk, and network usage — and sends them to an external API server.  
In addition, it monitors specified log files in real time and sends them in bulk, depending on configurable rules.

Designed for production environments where clarity and minimal overhead matter.

👉 [한글 버전](./README.ko.md)

---

## ✨ Key Features

| Feature                 | Description                                                                                   |
|-------------------------|-----------------------------------------------------------------------------------------------|
| Metric Reporting        | CPU, memory, disk, network, ports, and host info are periodically collected and sent|
| Docker & Redis Support  | Collects basic Docker container metadata and Redis server statistics (optional)|
| Log Collection          | Realtime log tailing with configurable buffer and flush timing|
| Retry & Backup          | Failed transmissions are queued and saved to file if exceeded retry count|
| Config-driven Behavior  | Controlled entirely via `config.yaml`, no code change required|
| UUID Assignment         | Each agent is assigned a persistent unique ID on first launch|
| Optional GET Endpoint   | When enabled, provides local `/metrics` endpoint for direct inspection|

---

## 🔐 Authentication

All requests from the agent include the following HTTP header:

```
Authorization: Bearer <auth_key>
```

This key is defined under `api.auth_key` in your config file.

---

## 🛠 Sample Configuration (`config.yaml`)

```yaml
uuid: "Agent-01"
interval: 10
retry_count: 300

http_server:
  enabled: true
  port: 6060

api:
  server: "http://localhost:5050/api/metrics"         # Leave empty to disable metrics Push
  heartbeat: "http://localhost:5050/api/heartbeat"    # Leave empty to disable heartbeat
  log: "http://localhost:5050/api/logs"               # Disabled if collectors.log.enabled is false
  auth_key: "test-secret"

collectors:
  cpu:
    enabled: true
  mem:
    enabled: true
  disk:
    enabled: true
  net:
    enabled: true
  ports:
    enabled: true
  host:
    enabled: true
  docker:
    enabled: true
  redis:
    enabled: true
    addr: "localhost:6379"
  log:
    enabled: true
    buffer_count: 5
    flush_interval: 20
    files:
      - C:\test\logs\test.log
      - /Users/test/Documents/revnoa/log/revnoa_2025-01-01.log

storage:
  file_backup:
    enabled: true
    dir: "/Users/test/Documents/backup/"

```

---

## 📦 Example Response

```json
{
  "timestamp": 1748437647,
  "data": {
    "agent_id": "Agent-01",
    "cpu": {
      "time_user_seconds": 7761.12,
      "time_system_seconds": 3084.39,
      "time_idle_seconds": 59606.99,
      "usage_percent": 1.74,
      "cores": 8,
      "load_1min": 2.90,
      "load_5min": 3.27,
      "load_15min": 3.16
    },
    "memory": {
      "total": 8589934592,
      "used": 6041841664,
      "free": 2548092928,
      "used_percent": 70.33,
      "swap_total": 1073741824,
      "swap_used": 126091264,
      "swap_used_percent": 11.74
    },
    "disks": [
      {
        "mount_point": "/",
        "total": 250685575168,
        "used": 46456938496,
        "used_perc": 18.53
      }
    ],
    "net": {
      "bytes_sent": 44623364,
      "bytes_recv": 413466388,
      "packets_sent": 207646,
      "packets_recv": 527650
    },
    "ports": [
      { "port": "51962" },
      { "port": "7000" },
      { "port": "5000" },
      { "port": "8080" },
      { "port": "5432" },
      { "port": "3306" }
    ],
    "host": {
      "hostname": "User-MacBookPro.local",
      "uptime": 10000,
      "os": "darwin",
      "platform": "darwin"
    },
    "docker": [
      {
        "id": "83c3e2e7b893",
        "name": "/test-nginx",
        "image": "nginx:latest",
        "status": "Up 3 hours"
      }
    ],
    "redis": {
      "status": "ok",
      "redis_version": "7.0.11",
      "redis_uptime_in_seconds": "4727",
      "redis_connected_clients": "2"
    }
  }
}
```

#### 📏 Units used in the response:

- `time_user_seconds`, `time_system_seconds`, `time_idle_seconds` → **seconds**
- `usage_percent`, `used_percent`, `swap_used_percent`, `load_1min`, `load_5min`, `load_15min` → **percent (%)**
- `total`, `used`, `free`, `swap_total`, `swap_used` → **bytes**
- `bytes_sent`, `bytes_recv` → **bytes**
- `packets_sent`, `packets_recv` → **count**
- `uptime` → **seconds**

#### 🔎 Note:

- **CPU load** (`load_1min`, etc.) represents the **average number of runnable processes**, not a percentage.
- **Redis values** are parsed directly from `INFO` and may remain as **strings**.
- **usage_percent** reflects the average CPU usage **between the current and previous collection** interval.

---

## 📎 Requirements

- Go 1.20+
- Docker API socket open (optional for Docker monitoring)
- Redis server available (optional for Redis metrics)

---

## 🔍 Notes

- Uses [`shirou/gopsutil`](https://github.com/shirou/gopsutil) for system metrics
- Uses [`nxadm/tail`](https://github.com/nxadm/tail) for log monitoring
- All background tasks support graceful shutdown on exit signals
- HMAC + timestamp-based request signing

---

## 🧪 Planned Improvements

This project currently collects metrics on demand whenever the `/metrics` endpoint is called.

In the future, I plan to support a buffer-based pull approach —  
where the agent collects metrics periodically in the background and stores them temporarily in memory.  
That way, when a request comes in, it can return the latest batch right away without needing to collect in real-time.

This is a small improvement that could help reduce response delay and better support polling from external systems.

---

## 🧭 License

MIT License — see `LICENSE` file.