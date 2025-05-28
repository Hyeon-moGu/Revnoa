package sender

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"revnoa/utils"
	"time"
)

var ApiKey = ""

func SetApiKey(apiKey string) {
	ApiKey = apiKey
}

func generateHMACSignature(body []byte, timestamp, apiKey string) string {
	dataToSign := string(body) + timestamp
	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(dataToSign))
	return hex.EncodeToString(mac.Sum(nil))
}

func SendPOST(url string, payload any, apiKey string, loggerPrefix string) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal failed: %v", err)
	}

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	signature := generateHMACSignature(body, timestamp, apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("non-2xx response: %d", resp.StatusCode)
	}

	utils.InfoLogger.Printf("[%s] sent successfully", loggerPrefix)
	return nil
}
