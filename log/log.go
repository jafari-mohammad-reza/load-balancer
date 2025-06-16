package log

import (
	"encoding/json"
	"errors"
	"fmt"
	"load-balancer/conf"
	"os"
	"sync"
	"time"
)

type ILogger interface {
	Info(args ...any) error
	Warn(args ...any) error
	Error(args ...any) error
	Debug(args ...any) error
}

func NewLogger(conf *conf.Conf) (ILogger, error) {
	switch conf.Log.Logger {
	case "json":
		return NewJsonLogger(conf), nil
	default:
		return nil, errors.New("unexpected logger")
	}
}

type JsonLogger struct {
	conf *conf.Conf
	mu   sync.Mutex
}

func NewJsonLogger(conf *conf.Conf) *JsonLogger {
	return &JsonLogger{
		conf: conf,
		mu:   sync.Mutex{},
	}
}

type LogLevel string

const (
	Info  LogLevel = "info"
	Warn  LogLevel = "warn"
	Error LogLevel = "error"
	Debug LogLevel = "debug"
)

type Log struct {
	DateTime string   `json:"date-time"`
	Level    LogLevel `json:"log-level"`
	Args     any      `json:"args"`
}

func (j *JsonLogger) write(level LogLevel, args ...any) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	entry := Log{
		DateTime: time.Now().Format(time.RFC3339Nano),
		Level:    level,
		Args:     toStringSlice(args),
	}

	var logs []Log
	data, err := os.ReadFile(j.conf.Log.LogPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read log file: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &logs); err != nil {
			return fmt.Errorf("failed to unmarshal existing log array: %w", err)
		}
	}

	logs = append(logs, entry)

	file, err := os.OpenFile(j.conf.Log.LogPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(logs); err != nil {
		return fmt.Errorf("failed to write updated log array: %w", err)
	}

	return nil
}
func toStringSlice(args []any) []string {
	result := make([]string, len(args))
	for i, arg := range args {
		result[i] = fmt.Sprint(arg)
	}
	return result
}

func (j *JsonLogger) Info(args ...any) error {
	return j.write(Info, args...)
}

func (j *JsonLogger) Warn(args ...any) error {
	return j.write(Warn, args...)
}

func (j *JsonLogger) Error(args ...any) error {
	return j.write(Error, args...)
}

func (j *JsonLogger) Debug(args ...any) error {
	return j.write(Debug, args...)
}
