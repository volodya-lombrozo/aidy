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

func (z *Zerolog) Info(args ...any) {
	z.logger.Info().Msgf("%v", args...)
}

func (z *Zerolog) Debug(args ...any) {
	z.logger.Debug().Msgf("%v", args...)
}

func (z *Zerolog) Warn(args ...any) {
	z.logger.Warn().Msgf("%v", args...)
}
