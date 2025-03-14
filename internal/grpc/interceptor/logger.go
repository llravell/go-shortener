package interceptor

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
)

func Logger(log *zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		log := log.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			log.Debug().Msg(msg)
		case logging.LevelInfo:
			log.Info().Msg(msg)
		case logging.LevelWarn:
			log.Warn().Msg(msg)
		case logging.LevelError:
			log.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
