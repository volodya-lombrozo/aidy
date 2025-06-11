package log

type Silent struct {
}

func NewSilent() Logger {
	return &Silent{}
}

func (s *Silent) Info(args ...any) {
}

func (s *Silent) Debug(args ...any) {
}

func (s *Silent) Warn(args ...any) {
}
