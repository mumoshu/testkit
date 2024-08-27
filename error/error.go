package error

import (
	"fmt"
	"strings"
)

// E is an error type with diagnostic information.
// It is inspired by Terraform's Diagnostic type.
//
// Unlike Diagnostic, E has no Severity field.
// That's because in testkit, whether it's an Error or a Warning
// depends on the context.
//
// In testkit apps, you combine E with log.L and appropriate Level
// so that the error and diagnostic information are logged
// according to the log level configured at runtime.
type E struct {
	// Cause is the cause of the error, if any.
	Cause error
	// Short description of the error.
	Short string
	// Source is the source of the error, if any.
	Source *source
	// Long description of the error.
	Long string
	// Remediation is the remediation for the error, if any.
	Remediation string
}

func (e E) writeS(b *strings.Builder, msg ...interface{}) {
	b.WriteString("╷\n")
	for i, s := range msg {
		if s == "" {
			continue
		}

		switch s := s.(type) {
		case string:
			if i > 0 {
				b.WriteString("│\n")
			}

			b.WriteString("│ ")
			b.WriteString(s)
			b.WriteString("\n")
		case *source:
			if s == nil {
				continue
			}

			if i > 0 {
				b.WriteString("│\n")
			}

			for _, l := range s.Lines() {
				b.WriteString("│   ")
				b.WriteString(l)
				b.WriteString("\n")
			}
		default:
			panic(fmt.Sprintf("unsupported type: %T", s))
		}
	}
	b.WriteString("╵")
}

func (e *E) String() string {
	var b strings.Builder

	e.writeS(&b,
		e.Short,
		e.Source,
		e.Long,
		e.Remediation,
	)

	return b.String()
}

func (e E) Error() string {
	return e.Short
}

type Option func(*E)

func Cause(err error) Option {
	return func(e *E) {
		e.Cause = err
	}
}

func Long(s string) Option {
	return func(e *E) {
		e.Long = s
	}
}

func Remediation(s string) Option {
	return func(e *E) {
		e.Remediation = s
	}
}

type source struct {
	FilePath string
	LineNum  int
	LineText string
	Group    string
}

func (s source) Lines() []string {
	var h string
	if s.Group == "" {
		h = fmt.Sprintf("on %s line %d:", s.FilePath, s.LineNum)
	} else {
		h = fmt.Sprintf("on %s line %d, in %s:", s.FilePath, s.LineNum, s.Group)

	}
	return []string{
		h,
		fmt.Sprintf("%2d: %s", s.LineNum, s.LineText),
	}
}

func Source(filePath string, lineNum int, lineText string) Option {
	return func(e *E) {
		e.Source = &source{
			FilePath: filePath,
			LineNum:  lineNum,
			LineText: lineText,
		}
	}
}

// New creates a new error.
func New(short string, opts ...Option) *E {
	e := &E{
		Short: short,
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Wrap wraps an error with a new error.
func Wrap(err error, short string, opts ...Option) E {
	e := E{
		Cause: err,
		Short: short,
	}

	for _, opt := range opts {
		opt(&e)
	}

	return e
}

func (e E) Unwrap() error {
	return e.Cause
}
