package log

import "os"

type Logger interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

var main Logger = NewShort(NewZerolog(os.Stdout, "debug"))

func Set(logger Logger) {
	if logger == nil {
		panic("logger cannot be nil")
	}
	main = logger
}

func Default() Logger {
	if main == nil {
		panic("logger is not set")
	}
	return main
}
