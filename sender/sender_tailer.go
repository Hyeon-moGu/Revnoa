package sender

import (
	"revnoa/utils"
)

type LogPayload struct {
	AgentID string   `json:"agent_id"`
	Lines   []string `json:"lines"`
}

func SendLogs(endpoint string, lines []string, agentID string) {
	if endpoint == "" {
		utils.WarnLogger.Println("Tailer Endpoint is empty")
		return
	}

	payload := LogPayload{
		AgentID: agentID,
		Lines:   lines,
	}

	utils.InfoLogger.Printf("Sending %d lines to %s", len(lines), endpoint)

	if err := SendPOST(endpoint, payload, ApiKey, "log"); err != nil {
		utils.ErrorLogger.Printf("Failed to send tailer: %v", err)
	}
}
