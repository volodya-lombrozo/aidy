package log

import (
	"fmt"
	"strings"
)

type Short struct {
	Length   int
	Original Logger
}

func NewShort(original Logger) Logger {
	return &Short{
		Length:   120,
		Original: original,
	}
}

func (s *Short) Debug(msg string, args ...any) {
	s.Original.Debug(s.logShortf(msg, args...))
}

func (s *Short) Info(msg string, args ...any) {
	s.Original.Info(s.logShortf(msg, args...))
}

func (s *Short) Warn(msg string, args ...any) {
	s.Original.Warn(s.logShortf(msg, args...))
}

func (s *Short) logShortf(format string, args ...any) string {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	if len(msg) > s.Length {
		msg = msg[:s.Length] + "…"
	}
	return strings.ReplaceAll(msg, "\n", " ")
}
