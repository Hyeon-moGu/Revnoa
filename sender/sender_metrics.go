package sender

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"revnoa/collector"
	"revnoa/utils"
	"time"
)

var (
	maxAllowedQueueSize = 300
	retryQueue          []MetricPayload

	maxRetryQueueSize = 100
	UseFileBackup     = true
	FileBackupDir     = "./log"
)

func SetBackup(count int) {
	maxRetryQueueSize = count
}

func SetStorageConfig(userFileBackUp bool, fileBackupDir string) {
	UseFileBackup = userFileBackUp
	FileBackupDir = fileBackupDir
}

type MetricPayload struct {
	Timestamp int64                 `json:"timestamp"`
	Data      collector.FullMetrics `json:"data"`
}

func generateBackupFilename() string {
	t := time.Now()
	return fmt.Sprintf("retry_backup_%s.json", t.Format("20060102_150405"))
}

func SendMetricsLoop(url string, current collector.FullMetrics) {
	var newRetryQueue []MetricPayload

	for _, item := range retryQueue {
		err := SendPOST(url, item.Data, ApiKey, "metrics")
		if err != nil {
			utils.ErrorLogger.Printf("Retry failed: %d", item.Timestamp)
			newRetryQueue = append(newRetryQueue, item)
		} else {
			utils.InfoLogger.Printf("Retry success: %d", item.Timestamp)
		}
	}

	payload := MetricPayload{
		Timestamp: current.Timestamp,
		Data:      current,
	}
	err := SendPOST(url, payload.Data, ApiKey, "metrics")
	if err != nil {
		utils.ErrorLogger.Printf("Send failed: %d", payload.Timestamp)
		newRetryQueue = append(newRetryQueue, payload)
	} else {
		utils.InfoLogger.Printf("Send success: %d", payload.Timestamp)
	}

	if len(newRetryQueue) >= maxRetryQueueSize {
		if UseFileBackup {
			_ = os.MkdirAll(FileBackupDir, 0755)
			fullPath := filepath.Join(FileBackupDir, generateBackupFilename())
			data, _ := json.MarshalIndent(newRetryQueue, "", "  ")
			if err := os.WriteFile(fullPath, data, 0644); err != nil {
				utils.ErrorLogger.Printf("Failed to save retry backup file: %v", err)
				retryQueue = newRetryQueue
			} else {
				utils.ErrorLogger.Printf("Retry queue overflow (%d), saved to: %s", len(newRetryQueue), fullPath)
				retryQueue = nil
			}
		} else {
			retryQueue = newRetryQueue
		}

		if len(newRetryQueue) > maxAllowedQueueSize {
			utils.WarnLogger.Printf("Queue size exceeded (%d), trimming to %d", len(newRetryQueue), maxAllowedQueueSize)
			newRetryQueue = newRetryQueue[len(newRetryQueue)-maxAllowedQueueSize:]
		}

		if retryQueue != nil {
			retryQueue = newRetryQueue
		}

	} else {
		retryQueue = newRetryQueue
	}
}
