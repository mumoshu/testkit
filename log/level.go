package log

import "fmt"

// Level is enum for log levels.
//
// The larger the value, the more verbose the log.
// For example, if you want to print something
// only when the configured log level is "info" or higher,
// you can write:
//
//	if configuedLogLevel >= log.Info {
//	  // Printed only when the configured log level is "info" or "debug"
//	  fmt.Println("something")
//	}
type Level int8

const (
	// Error is the error log level.
	Error Level = iota - 1
	// Info is the info log level.
	Info
	// Debug is the debug log level.
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
