package log

import (
	"fmt"
	"os"
	"sync"
)

// Printer is the interface for loggers to print logs.
// It can be anything that implements Printf(string, ...interface{}).
type Printer interface {
	Printf(format string, a ...interface{})
}

type defaultPrinter struct {
	Printer
	mu sync.Mutex
}

func (d *defaultPrinter) Printf(format string, a ...interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Printer == nil {
		d.Printer = PrinterFunc(func(f string, a ...interface{}) {
			_, err := fmt.Fprintf(os.Stderr, f+"\n", a...)
			if err != nil {
				panic(fmt.Errorf("failed to write to stderr: %v", err))
			}
		})
	}

	d.Printer.Printf(format, a...)
}
