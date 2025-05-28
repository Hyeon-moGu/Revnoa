package collector_test

import (
	"os"
	"testing"
	"time"

	"revnoa/collector"
	"revnoa/utils"
)

func TestTailerBasic(t *testing.T) {
	utils.InitLogger(true)

	tmpFile, err := os.CreateTemp("", "testlog-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	linesCollected := []string{}
	done := make(chan struct{})

	tailer := collector.NewTailer([]string{tmpFile.Name()}, 2, 2, func(lines []string) {
		linesCollected = append(linesCollected, lines...)
		if len(linesCollected) >= 2 {
			close(done)
		}
	})

	go tailer.Start()

	time.Sleep(500 * time.Millisecond)
	tmpFile.WriteString("First line\n")
	tmpFile.WriteString("Second line\n")
	tmpFile.Sync()

	select {
	case <-done:
		t.Logf("Collected lines: %v", linesCollected)
	case <-time.After(5 * time.Second):
		t.Fatal("Tailer did not collect expected lines")
	}

	tailer.Stop()
}
