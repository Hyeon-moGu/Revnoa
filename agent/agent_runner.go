package agent

import (
	"context"
	"fmt"
	"net/http"
	"revnoa/collector"
	"revnoa/config"
	"revnoa/handlers"
	"revnoa/sender"
	"revnoa/utils"
	"time"
)

var tailer collector.Tailer
var svr *http.Server

func RunAgent(ctx context.Context, cfg *config.Config, agentID string) {
	sender.SetBackup(cfg.RetryCount)
	sender.SetApiKey(cfg.API.AuthKey)
	sender.SetStorageConfig(cfg.Storage.FileBackup.Enabled, cfg.Storage.FileBackup.Dir)

	// Direct Health Check
	if cfg.API.Heartbeat != "" {
		sender.SendHealthLoop(agentID, cfg.API.Heartbeat)
	}

	// Metrics Collection Loop
	if cfg.API.Server != "" {
		go StartMetricsLoop(ctx, cfg, agentID)
	}

	// Log Tailer Task
	if cfg.Collectors.Log.Enabled {
		tailer = NewLogTailer(cfg, agentID)
		StartLogLoop(tailer)
	}

	// Metrics HTTP Endpoint
	if cfg.HTTPServer.Enabled {
		http.HandleFunc("/metrics", handlers.GetMetricsHandler(agentID, cfg))
		portStr := fmt.Sprintf(":%d", cfg.HTTPServer.Port)
		svr = &http.Server{Addr: portStr}

		utils.InfoLogger.Printf("Agent started - Push + Pull (port %d)\n", cfg.HTTPServer.Port)

		go func() {
			if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				utils.ErrorLogger.Fatalf("Failed to start HTTP server: %v", err)
			}
		}()
	} else {
		utils.InfoLogger.Println("Metrics GET endpoint is disabled")
	}

	// Graceful Shutdown
	<-ctx.Done()
	utils.InfoLogger.Printf("Shutting down agent gracefully at %s", time.Now().Format(time.RFC3339))

	if tailer != nil {
		tailer.Stop()
		utils.InfoLogger.Println("Tailer stopped")
	}

	if svr != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := svr.Shutdown(shutdownCtx); err != nil {
			utils.ErrorLogger.Printf("HTTP server shutdown error: %v", err)
		} else {
			utils.InfoLogger.Println("HTTP server stopped")
		}
	}
}
