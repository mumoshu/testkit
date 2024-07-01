package log

import (
	"runtime"
)

func New() *Leveled {
	return &Leveled{}
}

var Default = New()

type L struct {
	Logger
}

func (l *L) Debugf(format string, a ...interface{}) {
	l.get().Debugf(format, a...)
}

func (l *L) Infof(format string, a ...interface{}) {
	l.get().Infof(format, a...)
}

func (l *L) Errorf(format string, a ...interface{}) {
	l.get().Debugf(format, a...)
}

func (l *L) get() Logger {
	log := l.Logger

	if log == nil {
		pc, _, _, _ := runtime.Caller(2)
		f := runtime.FuncForPC(pc)
		name := f.Name()
		// i := strings.LastIndex(name, ".")
		// if i >= 0 {
		// 	name = name[:i]
		// }
		log = Default.For(name)
	}

	return log
}
