package config

import (
	"fmt"
	"os"
	"path/filepath"
	"revnoa/utils"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type Config struct {
	UUID       string        `yaml:"uuid"`
	Interval   int           `yaml:"interval"`
	RetryCount int           `yaml:"retry_count"`
	API        APIConfig     `yaml:"api"`
	Collectors CollectorSet  `yaml:"collectors"`
	Storage    StorageConfig `yaml:"storage"`
	HTTPServer HTTPServer    `yaml:"http_server"`
}

type APIConfig struct {
	Server    string `yaml:"server"`
	Heartbeat string `yaml:"heartbeat"`
	Log       string `yaml:"log"`
	AuthKey   string `yaml:"auth_key"`
}

type CollectorSet struct {
	CPU    CPUCollector  `yaml:"cpu"`
	Mem    MemCollector  `yaml:"mem"`
	Net    GenericSwitch `yaml:"net"`
	Disk   GenericSwitch `yaml:"disk"`
	Ports  GenericSwitch `yaml:"ports"`
	Host   GenericSwitch `yaml:"host"`
	Docker GenericSwitch `yaml:"docker"`
	Redis  RedisConfig   `yaml:"redis"`
	Log    LogCollector  `yaml:"log"`
}

type GenericSwitch struct {
	Enabled bool `yaml:"enabled"`
}

type CPUCollector struct {
	Enabled bool `yaml:"enabled"`
}

type MemCollector struct {
	Enabled bool `yaml:"enabled"`
}

type RedisConfig struct {
	Enabled bool   `yaml:"enabled"`
	Addr    string `yaml:"addr"`
}

type LogCollector struct {
	Enabled       bool     `yaml:"enabled"`
	BufferCount   int      `yaml:"buffer_count"`
	FlushInterval int      `yaml:"flush_interval"`
	Files         []string `yaml:"files"`
}

type StorageConfig struct {
	FileBackup FileBackupConfig `yaml:"file_backup"`
}

type FileBackupConfig struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
}

type HTTPServer struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}

func LoadConfig(path string) (*Config, error) {
	cfg := &Config{}
	data, err := os.ReadFile(path)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to read config file: %v", err)
		return nil, err
	}

	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to parse config file: %v", err)
		return nil, err
	}

	for i, file := range cfg.Collectors.Log.Files {
		cfg.Collectors.Log.Files[i] = filepath.Clean(file)
	}

	if cfg.UUID == "" {
		newUUID := uuid.New().String()
		cfg.UUID = newUUID
		_ = injectUUIDToFile(path, newUUID)
		utils.InfoLogger.Printf("Generated UUID: %s", newUUID)
	}

	utils.InfoLogger.Println("Config loaded successfully")
	return cfg, nil
}

func (c *Config) Validate() error {
	var errs []string

	if strings.TrimSpace(c.UUID) == "" {
		errs = append(errs, "UUID must not be empty")
	}
	if strings.TrimSpace(c.API.Log) == "" {
		errs = append(errs, "API.Log must not be empty")
	}
	if c.Interval <= 0 {
		errs = append(errs, "Interval must be greater than 0")
	}
	if c.RetryCount < 0 {
		errs = append(errs, "RetryCount must be non-negative")
	}
	if c.Collectors.Log.Enabled {
		if c.Collectors.Log.BufferCount <= 0 {
			errs = append(errs, "Log buffer_count must be > 0")
		}
		if c.Collectors.Log.FlushInterval <= 0 {
			errs = append(errs, "Log flush_interval must be > 0")
		}
		if len(c.Collectors.Log.Files) == 0 {
			errs = append(errs, "Log files must include at least one path")
		}
	}
	if c.Storage.FileBackup.Enabled && strings.TrimSpace(c.Storage.FileBackup.Dir) == "" {
		utils.WarnLogger.Println("File backup is enabled but dir is empty")
	}

	if c.Collectors.Redis.Enabled && strings.TrimSpace(c.Collectors.Redis.Addr) == "" {
		errs = append(errs, "Redis address must be set if redis is enabled")
	}

	if c.HTTPServer.Enabled && c.HTTPServer.Port <= 0 {
		errs = append(errs, "HTTP server port must be > 0 if enabled")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

func injectUUIDToFile(path, newUUID string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to read config for UUID injection: %v", err)
		return err
	}

	lines := strings.Split(string(content), "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "uuid:") {
			lines[i] = fmt.Sprintf("uuid: \"%s\"", newUUID)
			found = true
			break
		}
	}

	if !found {
		lines = append([]string{fmt.Sprintf("uuid: \"%s\"", newUUID)}, lines...)
	}

	err = os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	if err != nil {
		utils.ErrorLogger.Printf("Failed to write UUID to config: %v", err)
	}
	return err
}
