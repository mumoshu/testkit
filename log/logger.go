package log

type Logger interface {
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Errorf(format string, a ...interface{})
}

type namedLeveled struct {
	name string
	l    *Leveled
}

func (n *namedLeveled) Debugf(format string, a ...interface{}) {
	n.l.Logf(n.name, Debug, format, a...)
}

func (n *namedLeveled) Infof(format string, a ...interface{}) {
	n.l.Logf(n.name, Info, format, a...)
}

func (n *namedLeveled) Errorf(format string, a ...interface{}) {
	n.l.Logf(n.name, Error, format, a...)
}

type leveled struct {
	l *Leveled
}

func (l *leveled) Debugf(format string, a ...interface{}) {
	l.l.Logf("", Debug, format, a...)
}

func (l *leveled) Infof(format string, a ...interface{}) {
	l.l.Logf("", Info, format, a...)
}

func (l *leveled) Errorf(format string, a ...interface{}) {
	l.l.Logf("", Error, format, a...)
}
