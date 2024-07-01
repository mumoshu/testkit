package log

import "fmt"

// Level is enum for log levels.
type Level int8

const (
	// Debug is the debug log level.
	Error Level = iota - 1
	// Info is the info log level.
	Info
	// Error is the error log level.
	Debug
)

func (l Level) String() string {
	switch l {
	case Error:
		return "error"
	case Info:
		return "info"
	case Debug:
		return "debug"
	default:
		return fmt.Sprintf("(unknown log level: %d)", l)
	}
}
