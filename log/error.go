package log

import (
	testkiterror "github.com/mumoshu/testkit/error"
)

// E prints the error at the specified severity using the leveled logger l.
//
// Similarly to log.Leveled is, this is intended for use in the library code,
// rather than the application code.
// An application should use Logger.E instead.
func E(l *Leveled, name string, severity Level, e error) {
	var msg string

	if e2, ok := e.(*testkiterror.E); ok {
		// Print the full message
		// only when the log level is configured to be more verbose
		// than the severity of the error.
		//
		// For example, if the log level is "info",
		// only print the full message when the severity is "info" or "error".
		if l.Level >= severity {
			msg = e2.String()
		} else {
			msg = e2.Short
		}
	} else {
		msg = e.Error()
	}

	l.Logf(name, severity, msg)
}
