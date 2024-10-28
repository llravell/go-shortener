package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestDecompressMiddleware(t *testing.T) {
	router := chi.NewRouter()

	router.Use(DecompressMiddleware())
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	testCases := []struct {
		name         string
		body         io.Reader
		headers      map[string]string
		expectedBody string
	}{
		{
			name:         "Client send data without compress",
			body:         strings.NewReader("plain"),
			expectedBody: "plain",
		},
		{
			name: "Client send data with compress",
			body: compressReader(t, "compressed"),
			headers: map[string]string{
				"Content-Encoding": "gzip",
			},
			expectedBody: "compressed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, body := testRequest(t, ts, http.MethodPost, "/", tc.body, tc.headers)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedBody, string(body))
		})
	}
}
