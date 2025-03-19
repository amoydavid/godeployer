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
	"time"

	"github.com/fatih/color"
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

// Logger 是一个简单的日志记录器
type Logger struct {
	writer  io.Writer
	colored bool
	level   LogLevel
}

// NewLogger 创建一个新的日志记录器
func NewLogger(writer io.Writer, colored bool) *Logger {
	return &Logger{
		writer:  writer,
		colored: colored,
		level:   INFO,
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// log 记录一条消息
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	var prefix string

	if l.colored {
		switch level {
		case DEBUG:
			prefix = color.BlueString("[DEBUG]")
		case INFO:
			prefix = color.CyanString("[INFO]")
		case SUCCESS:
			prefix = color.GreenString("[SUCCESS]")
		case WARN:
			prefix = color.YellowString("[WARN]")
		case ERROR:
			prefix = color.RedString("[ERROR]")
		}
	} else {
		switch level {
		case DEBUG:
			prefix = "[DEBUG]"
		case INFO:
			prefix = "[INFO]"
		case SUCCESS:
			prefix = "[SUCCESS]"
		case WARN:
			prefix = "[WARN]"
		case ERROR:
			prefix = "[ERROR]"
		}
	}

	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(l.writer, "%s %s %s\n", timestamp, prefix, message)
}

// Debug 记录一条调试消息
func (l *Logger) Debug(message string) {
	l.log(DEBUG, message)
}

// Debugf 使用格式记录一条调试消息
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 记录一条信息消息
func (l *Logger) Info(message string) {
	l.log(INFO, message)
}

// Infof 使用格式记录一条信息消息
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Success 记录一条成功消息
func (l *Logger) Success(message string) {
	l.log(SUCCESS, message)
}

// Successf 使用格式记录一条成功消息
func (l *Logger) Successf(format string, args ...interface{}) {
	l.log(SUCCESS, format, args...)
}

// Warn 记录一条警告消息
func (l *Logger) Warn(message string) {
	l.log(WARN, message)
}

// Warnf 使用格式记录一条警告消息
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 记录一条错误消息
func (l *Logger) Error(message string) {
	l.log(ERROR, message)
}

// Errorf 使用格式记录一条错误消息
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}
