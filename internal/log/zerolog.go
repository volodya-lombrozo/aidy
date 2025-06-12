package log

import (
	"io"

	"github.com/rs/zerolog"
)

type Zerolog struct {
	logger zerolog.Logger
}

func NewZerolog(writer io.Writer) Logger {
	return &Zerolog{
		logger: zerolog.New(zerolog.ConsoleWriter{Out: writer}).With().Timestamp().Logger(),
	}
}

func (z *Zerolog) Info(msg string, args ...any) {
	z.logger.Info().Msgf(msg, args...)
}

func (z *Zerolog) Debug(msg string, args ...any) {
	z.logger.Debug().Msgf(msg, args...)
}

func (z *Zerolog) Warn(msg string, args ...any) {
	z.logger.Warn().Msgf(msg, args...)
}

func (z *Zerolog) Error(msg string, args ...any) {
	z.logger.Error().Msgf(msg, args...)
}
