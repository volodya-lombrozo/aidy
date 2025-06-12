package log

type Logger interface {
	Info(string, ...any)
	Debug(string, ...any)
	Warn(string, ...any)
}
