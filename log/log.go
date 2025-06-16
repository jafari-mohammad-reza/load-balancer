package log

import (
	"errors"
	"load-balancer/conf"
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
