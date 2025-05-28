package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	WarnLogger  *log.Logger
)

func InitLogger(disableFile bool) {
	var multiOut io.Writer

	if disableFile {
		multiOut = os.Stdout
	} else {
		t := time.Now().Format("2006-01-02")
		logDir := "revnoa_log"
		_ = os.MkdirAll(logDir, 0755)

		filePath := filepath.Join(logDir, fmt.Sprintf("revnoa_%s.log", t))
		file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Make Dir failed: %v", err)
		}
		multiOut = io.MultiWriter(file, os.Stdout)
	}

	InfoLogger = log.New(multiOut, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(multiOut, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)
	WarnLogger = log.New(multiOut, "[WARN] ", log.Ldate|log.Ltime|log.Lshortfile)
}
