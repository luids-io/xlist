// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. See LICENSE.

package loggerwr

import (
	"errors"
	"strings"
)

// LogLevel defines LogLevel for rules
type LogLevel int

// Values for LogLevel
const (
	Disable LogLevel = iota
	Debug
	Info
	Warn
	Error
)

// StringToLevel returns the loglevel
func StringToLevel(s string) (LogLevel, error) {
	switch strings.ToLower(s) {
	case "disable":
		return Disable, nil
	case "debug":
		return Debug, nil
	case "info":
		return Info, nil
	case "warn":
		return Warn, nil
	case "error":
		return Error, nil
	default:
		return Disable, errors.New("invalid level")
	}
}

// Logger defines the interface for the logger
type Logger interface {
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}
