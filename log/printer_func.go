package log

type PrinterFunc func(format string, a ...interface{})

var _ Printer = PrinterFunc(nil)

func (f PrinterFunc) Printf(format string, a ...interface{}) {
	f(format, a...)
}
