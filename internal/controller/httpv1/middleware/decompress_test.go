package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/stretchr/testify/assert"
)

func TestDecompressMiddleware(t *testing.T) {
	router := chi.NewRouter()

	router.Use(middleware.DecompressMiddleware())
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	testCases := []struct {
		name           string
		body           io.Reader
		headers        map[string]string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Client send data without compress",
			body:           strings.NewReader("plain"),
			expectedBody:   "plain",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Client send data with compress",
			body: compressReader(t, "compressed"),
			headers: map[string]string{
				"Content-Encoding": "gzip",
			},
			expectedBody:   "compressed",
			expectedStatus: http.StatusOK,
		},
		{
			name: "Client send data with incorrect encoding header",
			body: strings.NewReader("not compressed"),
			headers: map[string]string{
				"Content-Encoding": "gzip",
			},
			expectedBody:   "gzip decoding error\n",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, body := testutils.SendTestRequest(t, ts, ts.Client(), http.MethodPost, "/", tc.body, tc.headers)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)
			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
