package log

import (
	"fmt"
)

// Env is the default name of the environment variable that contains the log entries.
var Env = "TESTKIT_LOG"

type Leveled struct {
	Debug, Info, Error Printer

	defaultPrinter

	leveledCache
}

func (l *Leveled) Logf(name string, level Level, format string, a ...interface{}) {
	if !l.Enabled(name, level) {
		return
	}

	var printer Printer

	switch level {
	case Debug:
		printer = l.Debug
	case Info:
		printer = l.Info
	case Error:
		printer = l.Error
	default:
		panic(fmt.Errorf("unknown log level: %v", level))
	}

	format = "[%s %s] " + format
	args := append([]interface{}{name, level}, a...)

	if printer == nil {
		l.Printf(format, args...)
		return
	}

	printer.Printf(format, args...)
}

func (l *Leveled) For(name string) *namedLeveled {
	return &namedLeveled{
		name: name,
		l:    l,
	}
}
