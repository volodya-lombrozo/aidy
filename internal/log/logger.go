package log

type Logger interface {
	Info(args ...any)
	Debug(args ...any)
	Warn(args ...any)
}
