package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

type logFormatter struct {
	logger zerolog.Logger
}

func (l *logFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &logEntry{
		logFormatter: l,
		request:      r,
	}
}

type logEntry struct {
	*logFormatter
	request *http.Request
}

func (l *logEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ interface{}) {
	l.logger.Info().
		Str("remote_addr", l.request.RemoteAddr).
		Str("method", l.request.Method).
		Str("uri", l.request.RequestURI).
		Int("status", status).
		Int("size", bytes).
		Dur("duration", elapsed).
		Msg("incoming request")
}

func (l *logEntry) Panic(v interface{}, _ []byte) {
	middleware.PrintPrettyStack(v)
}

func LoggerMiddleware(l zerolog.Logger) func(next http.Handler) http.Handler {
	lf := &logFormatter{logger: l}

	return middleware.RequestLogger(lf)
}
