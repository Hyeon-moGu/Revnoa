package agent

import (
	"os"
	"revnoa/collector"
	"revnoa/config"
	"revnoa/sender"
	"revnoa/utils"
)

func NewLogTailer(cfg *config.Config, agentID string) collector.Tailer {
	validFiles := make([]string, 0, len(cfg.Collectors.Log.Files))

	for _, path := range cfg.Collectors.Log.Files {
		if _, err := os.Stat(path); err == nil {
			validFiles = append(validFiles, path)
		} else {
			utils.WarnLogger.Printf("Can't read log file: %s", path)
		}
	}

	if len(validFiles) == 0 {
		utils.WarnLogger.Println("No log files found. Skip log tail.")
		return nil
	}

	return collector.NewTailer(
		validFiles,
		cfg.Collectors.Log.BufferCount,
		cfg.Collectors.Log.FlushInterval,
		func(lines []string) {
			sender.SendLogs(cfg.API.Log, lines, agentID)
		},
	)
}

func StartLogLoop(tailer collector.Tailer) {
	if tailer == nil {
		utils.WarnLogger.Println("No log tailer. Skip.")
		return
	}

	if err := tailer.Start(); err != nil {
		utils.ErrorLogger.Printf("Fail to start log tailer: %v", err)
	} else {
		utils.InfoLogger.Println("Log tailer started.")
	}
}
