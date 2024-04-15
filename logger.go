package tksql

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// LogLevel ログレベル.
type LogLevel int

const (
	LogLevelNone = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

var (
	// logger ロガー.
	logger     = &Logger{}
	moduleName string
)

// LogConfig ログ設定構造体.
type LogConfig struct {
	Level LogLevel
	Debug func(message string)
	Info  func(message string)
	Warn  func(message string)
	Error func(message string)
}

// Logger ロガー構造体.
type Logger struct {
	config LogConfig
}

// Debug デバッグログを出力します.
func (l *Logger) Debug(messages ...any) {
	if l.checkLevel(LogLevelDebug) {
		l.config.Debug(createMessage(messages))
	}
}

// Info 情報ログを出力します.
func (l *Logger) Info(messages ...any) {
	if l.checkLevel(LogLevelInfo) {
		l.config.Info(createMessage(messages))
	}
}

// Info 警告ログを出力します.
func (l *Logger) Warn(messages ...any) {
	if l.checkLevel(LogLevelWarn) {
		l.config.Warn(createMessage(messages))
	}
}

// Info 異常ログを出力します.
func (l *Logger) Error(messages ...any) {
	if l.checkLevel(LogLevelError) {
		l.config.Error(createMessage(messages))
	}
}

// IsDebagEnable デバッグログ出力が有効か判定します.
func (l *Logger) IsDebagEnable() bool {
	return l.checkLevel(LogLevelDebug)
}

// checkLevel 出力対象のログレベルか判定します.
func (l *Logger) checkLevel(level LogLevel) bool {
	return l.config.Level <= level
}

// createMessage ログメッセージを作成します.
func createMessage(messages ...any) string {
	message := fmt.Sprintf("%v", messages)
	message = strings.TrimPrefix(message, "[[")
	message = strings.TrimSuffix(message, "]]")

	if len(moduleName) == 0 {
		selfUnitPtr, _, _, _ := runtime.Caller(0)
		selfPCName := runtime.FuncForPC(selfUnitPtr).Name()
		moduleName = regexp.MustCompile(`(\.\(.+\))*[\./]createMessage`).ReplaceAllString(selfPCName, ".")
		fmt.Println(moduleName)
	}

	up, _, line, _ := runtime.Caller(2)
	pac := strings.Replace(runtime.FuncForPC(up).Name(), moduleName, "", 1)
	message = "[tksql - " + pac + ":" + strconv.Itoa(line) + "] " + message
	return message
}
