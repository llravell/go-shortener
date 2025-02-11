// Пакет logger конфигурирует логер и возвращает единственный инстанс
package logger

import (
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var once sync.Once

var log zerolog.Logger

var logFile *os.File

const (
	logFilePermissions = 0o664
	logSample          = 10
)

// Get возвращает сгенерированный логгер. Конфигурация происходит единожды.
func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.InfoLevel) // default to INFO
		}

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		appEnv := os.Getenv("APP_ENV")
		if appEnv != "development" {
			logFile, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, logFilePermissions)
			if err != nil {
				panic(err)
			}

			output = zerolog.MultiLevelWriter(os.Stdout, logFile)
		}

		var goVersion string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			goVersion = buildInfo.GoVersion
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)). //nolint:gosec // disable G115
			With().
			Timestamp().
			Str("go_version", goVersion).
			Str("env", appEnv).
			Logger().
			Sample(&zerolog.BasicSampler{N: logSample})
	})

	return log
}

// Close закрывает файл, если он использовался для записи логов.
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}
