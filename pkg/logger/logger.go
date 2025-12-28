/*
 * GoDeployer - 灵活、可扩展的部署工具
 * Copyright (C) 2025-2025 amoydavid
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel 表示日志级别
type LogLevel int

const (
	// DEBUG 级别
	DEBUG LogLevel = iota
	// INFO 级别
	INFO
	// SUCCESS 级别
	SUCCESS
	// WARN 级别
	WARN
	// ERROR 级别
	ERROR
)

// Logger 是一个结构化日志记录器，基于 zap
type Logger struct {
	zapLogger *zap.Logger
	sugar     *zap.SugaredLogger
	colored   bool
	level     LogLevel
}

// NewLogger 创建一个新的结构化日志记录器
func NewLogger(writer io.Writer, colored bool) *Logger {
	// Determine log level
	var zapLevel zapcore.Level
	zapLevel = zapcore.InfoLevel

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create console encoder for human-readable output
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Create writer (use provided writer or stdout)
	var writeSyncer zapcore.WriteSyncer
	if writer != nil {
		// Check if writer implements Sync() method
		type syncer interface {
			io.Writer
			Sync() error
		}
		if w, ok := writer.(syncer); ok {
			writeSyncer = zapcore.AddSync(w)
		} else {
			// Writer doesn't have Sync, wrap it
			writeSyncer = zapcore.Lock(zapcore.AddSync(writer))
		}
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Build core
	core := zapcore.NewCore(consoleEncoder, writeSyncer, zapLevel)

	// Create logger
	zapLogger := zap.New(core, zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		zapLogger: zapLogger,
		sugar:     zapLogger.Sugar(),
		colored:   colored,
		level:     INFO,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// WithFields returns a logger with structured fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	newSugar := l.sugar.With(args...)
	return &Logger{
		zapLogger: l.zapLogger,
		sugar:     newSugar,
		colored:   l.colored,
		level:     l.level,
	}
}

// formatWithColor formats message with color if enabled
func (l *Logger) formatWithColor(level LogLevel, message string) string {
	if !l.colored {
		return message
	}

	switch level {
	case DEBUG:
		return color.BlueString(message)
	case INFO:
		return color.CyanString(message)
	case SUCCESS:
		return color.GreenString(message)
	case WARN:
		return color.YellowString(message)
	case ERROR:
		return color.RedString(message)
	default:
		return message
	}
}

// log 记录一条结构化消息
func (l *Logger) log(level LogLevel, msg string, args ...interface{}) {
	if level < l.level {
		return
	}

	formattedMsg := l.formatWithColor(level, msg)

	switch level {
	case DEBUG:
		if len(args) == 0 {
			l.sugar.Debug(formattedMsg)
		} else {
			l.sugar.Debugf(formattedMsg, args...)
		}
	case INFO:
		if len(args) == 0 {
			l.sugar.Info(formattedMsg)
		} else {
			l.sugar.Infof(formattedMsg, args...)
		}
	case SUCCESS:
		// SUCCESS maps to INFO level in zap but with green color
		if len(args) == 0 {
			l.sugar.Info(formattedMsg)
		} else {
			l.sugar.Infof(formattedMsg, args...)
		}
	case WARN:
		if len(args) == 0 {
			l.sugar.Warn(formattedMsg)
		} else {
			l.sugar.Warnf(formattedMsg, args...)
		}
	case ERROR:
		if len(args) == 0 {
			l.sugar.Error(formattedMsg)
		} else {
			l.sugar.Errorf(formattedMsg, args...)
		}
	}
}

// Debug records a debug message
func (l *Logger) Debug(message string) {
	l.log(DEBUG, message)
}

// Debugf records a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...))
}

// Info records an info message
func (l *Logger) Info(message string) {
	l.log(INFO, message)
}

// Infof records a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...))
}

// Success records a success message
func (l *Logger) Success(message string) {
	l.log(SUCCESS, message)
}

// Successf records a formatted success message
func (l *Logger) Successf(format string, args ...interface{}) {
	l.log(SUCCESS, fmt.Sprintf(format, args...))
}

// Warn records a warning message
func (l *Logger) Warn(message string) {
	l.log(WARN, message)
}

// Warnf records a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...))
}

// Error records an error message
func (l *Logger) Error(message string) {
	l.log(ERROR, message)
}

// Errorf records a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...))
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.zapLogger.Sync()
}
