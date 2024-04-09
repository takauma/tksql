package tksql

// LogLevel ログレベル.
type LogLevel int

const (
	LogLevelNone = iota
	LogLevelDebug
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// LogConfig ログ設定構造体.
type LogConfig struct {
	Level LogLevel
	Debug func(message string)
	Info  func(message string)
	Warn  func(message string)
	Error func(message string)
}
