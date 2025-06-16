package log

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"load-balancer/conf"
)

func TestJsonLogger_WriteLogEntries(t *testing.T) {

	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test-log.json")

	cfg := &conf.Conf{
		Log: conf.LogConf{
			LogPath: logPath,
		},
	}

	logger := NewJsonLogger(cfg)

	if err := logger.Info("info log entry", 123); err != nil {
		t.Fatalf("failed to write info log: %v", err)
	}
	if err := logger.Warn("warn log entry"); err != nil {
		t.Fatalf("failed to write warn log: %v", err)
	}
	if err := logger.Error("error log entry"); err != nil {
		t.Fatalf("failed to write error log: %v", err)
	}
	if err := logger.Debug("debug log entry"); err != nil {
		t.Fatalf("failed to write debug log: %v", err)
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	var logs []Log
	if err := json.Unmarshal(data, &logs); err != nil {
		t.Fatalf("failed to unmarshal log file: %v", err)
	}

	if len(logs) != 4 {
		t.Fatalf("expected 4 log entries, got %d", len(logs))
	}

	expectedLevels := []LogLevel{Info, Warn, Error, Debug}
	expectedMessages := []string{
		"info log entry",
		"warn log entry",
		"error log entry",
		"debug log entry",
	}

	for i, entry := range logs {
		if entry.Level != expectedLevels[i] {
			t.Errorf("expected level %s, got %s", expectedLevels[i], entry.Level)
		}
		args, ok := entry.Args.([]interface{})
		if !ok {
			t.Errorf("entry.Args is not a slice, got %T", entry.Args)
			continue
		}
		combined := ""
		for _, arg := range args {
			str, ok := arg.(string)
			if !ok {
				t.Errorf("arg is not string, got %T", arg)
				continue
			}
			combined += str + " "
		}
		if combined[:len(expectedMessages[i])] != expectedMessages[i] {
			t.Errorf("log[%d] unexpected message: got %q, want %q", i, combined, expectedMessages[i])
		}
		if _, err := time.Parse(time.RFC3339Nano, entry.DateTime); err != nil {
			t.Errorf("invalid datetime format: %v", err)
		}
	}
}
