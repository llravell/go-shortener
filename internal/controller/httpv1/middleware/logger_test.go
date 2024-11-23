package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Log struct {
	Level    string  `json:"level"`
	Method   string  `json:"method"`
	URI      string  `json:"uri"`
	Status   int     `json:"status"`
	Size     int     `json:"size"`
	Duration float32 `json:"duration"`
	Message  string  `json:"message"`
}

func readLog(t *testing.T, r *bytes.Buffer) *Log {
	t.Helper()

	var log Log

	err := json.Unmarshal(r.Bytes(), &log)
	require.NoError(t, err)

	return &log
}

const payload = "echo"

func TestLoggerMiddleware(t *testing.T) {
	router := chi.NewRouter()

	out := &bytes.Buffer{}
	logger := zerolog.New(out)

	router.Use(middleware.LoggerMiddleware(logger))
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	t.Run("Test request log", func(t *testing.T) {
		out.Reset()

		res, _ := testutils.SendTestRequest(
			t, ts, ts.Client(), http.MethodPost, "/", strings.NewReader(payload), map[string]string{},
		)
		defer res.Body.Close()

		log := readLog(t, out)

		assert.Equal(t, "info", log.Level)
		assert.Equal(t, http.MethodPost, log.Method)
		assert.Equal(t, http.StatusOK, log.Status)
		assert.Equal(t, "incoming request", log.Message)
		assert.Equal(t, "/", log.URI)
		assert.Equal(t, len(payload), log.Size)
		assert.Greater(t, log.Duration, float32(0))
	})
}
