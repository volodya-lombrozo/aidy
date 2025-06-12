package log

type Silent struct {
}

func NewSilent() Logger {
	return &Silent{}
}

func (s *Silent) Info(msg string, args ...any) {
}

func (s *Silent) Debug(msg string, args ...any) {
}

func (s *Silent) Warn(msg string, args ...any) {
}

func (s *Silent) Error(msg string, args ...any) {
}
