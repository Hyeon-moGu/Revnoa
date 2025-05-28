package sender

import (
	"revnoa/utils"
	"time"
)

type Heartbeat struct {
	AgentID   string `json:"agent_id"`
	Timestamp int64  `json:"timestamp"`
}

func SendHealthLoop(agentID, apiUrl string) {
	go func() {
		for {
			payload := Heartbeat{
				AgentID:   agentID,
				Timestamp: time.Now().Unix(),
			}

			if err := SendPOST(apiUrl, payload, ApiKey, "heartbeat"); err != nil {
				utils.WarnLogger.Printf("Heartbeat send failed: %v", err)
			}

			time.Sleep(60 * time.Second)
		}
	}()
}
