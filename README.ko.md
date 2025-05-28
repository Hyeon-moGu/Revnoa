# Revnoa

**Revnoa**는 Go로 작성된 경량 시스템 모니터링 에이전트입니다.  
CPU, 메모리, 디스크, 네트워크 사용량 등 주요 시스템 메트릭을 주기적으로 수집하고 외부 API 서버로 전송합니다.  
또한 지정된 로그 파일을 실시간으로 모니터링하며, 설정한 조건에 따라 버퍼링하여 전송합니다.

👉 [English Version](./README.md)

---

## ✨ 주요 기능

| 기능                      | 설명                                                                 |
|---------------------------|----------------------------------------------------------------------|
| 시스템 메트릭 수집         | CPU, 메모리, 디스크, 네트워크, 포트, 호스트 정보 등을 주기적으로 수집 및 전송|
| Docker 및 Redis 수집 지원 | Docker 컨테이너 메타데이터 및 Redis 서버 정보 수집 기능 (옵션)|
| 로그 수집                 | 로그 파일을 tail 방식으로 실시간 모니터링, 버퍼 설정에 따라 묶어서 전송|
| 전송 실패 대응             | 실패 시 큐에 적재하며, 재시도 횟수 초과 시 파일로 백업 저장|
| 설정 기반 동작            | 모든 동작은 `config.yaml` 파일을 통해 설정|
| UUID 자동 생성             | 실행 시 고유 UUID 자동 생성 및 설정 파일에 반영|
| GET /metrics 지원         | 로컬 HTTP 서버를 통해 실시간 메트릭을 조회할 수 있는 /metrics 엔드포인트 제공 (옵션)|

---

## 🔐 인증 방식

에이전트는 모든 요청에 아래와 같은 HTTP 헤더를 포함

```
Authorization: Bearer <auth_key>
```

해당 키는 `config.yaml`의 `api.auth_key` 필드에 정의

---

## 🛠 설정 예시 (`config.yaml`)

```yaml
uuid: "Agent-01"
interval: 10
retry_count: 300

http_server:
  enabled: true
  port: 6060

api:
  server: "http://localhost:5050/api/metrics"         # 빈 값이면 metrics push 비활성화
  heartbeat: "http://localhost:5050/api/heartbeat"    # 빈 값이면 health push 비활성화
  log: "http://localhost:5050/api/logs"               # collectors.log.enabled=false시 비활성화
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

## 📦 메트릭 응답 예시

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

#### 📏 사용되는 단위

- `time_user_seconds`, `time_system_seconds`, `time_idle_seconds` → **초 (seconds)**
- `usage_percent`, `used_percent`, `swap_used_percent`, `load_1min`, `load_5min`, `load_15min` → **퍼센트 (%)**
- `total`, `used`, `free`, `swap_total`, `swap_used` → **바이트 (bytes)**
- `bytes_sent`, `bytes_recv` → **바이트 (bytes)**
- `packets_sent`, `packets_recv` → **개수 (count)**
- `uptime` → **초 (seconds)**

#### 🔎 참고사항

- **CPU 부하** (`load_1min` 등)는 **실행 대기 중인 프로세스 수 평균값**
- **Redis 값**은 `INFO` 명령 결과를 파싱한 것으로 **문자열로 유지**
- `usage_percent`는 **이전 수집 시점과 현재 시점 간 평균 CPU 사용률**을 의미

---

## 📎 필요 환경

- Go 1.20 이상
- Docker API 소켓 활성화 필요 (옵션)
- Redis 서버 접근 가능여부 (옵션)

---

## 🔍 기타

- 시스템 메트릭 수집: [`shirou/gopsutil`](https://github.com/shirou/gopsutil)
- 로그 모니터링: [`nxadm/tail`](https://github.com/nxadm/tail)
- 종료 시 모든 백그라운드 작업 안전 종료 처리
- 모든 요청은 HMAC 서명과 타임스탬프를 포함하여 위변조를 방지

---

## 🧪 개선 사항
현재 /metrics 엔드포인트가 호출될 때마다 실시간으로 메트릭을 수집하는 구조입니다.

향후에는 에이전트가 백그라운드에서 일정 주기로 메트릭을 수집하고,
그 데이터를 메모리에 임시로 저장해두는 버퍼 기반 Pull 방식으로 지원할 예정입니다.

이 방식은 수집 요청이 들어왔을 때 실시간 수집 없이 가장 최근 데이터를 바로 응답할 수 있어,
응답 지연을 줄이고 외부 시스템에서 주기적으로 수집(polling)하는 환경에서도 더 안정적으로 동작할 수 있습니다.

---

## 🧭 라이선스

MIT License — 자세한 내용은 `LICENSE` 참고
