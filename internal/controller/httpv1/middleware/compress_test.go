package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/stretchr/testify/assert"
)

func TestCompressMiddleware(t *testing.T) {
	router := chi.NewRouter()

	router.Use(middleware.CompressMiddleware("text/plain"))
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	testCases := []struct {
		name               string
		body               io.Reader
		headers            map[string]string
		expectedBody       string
		shouldBeCompressed bool
	}{
		{
			name:         "Server return data without compress",
			body:         strings.NewReader("plain"),
			expectedBody: "plain",
		},
		{
			name: "Server return data with compress",
			body: strings.NewReader("compressed"),
			headers: map[string]string{
				"Accept-Encoding": "gzip",
			},
			expectedBody:       "compressed",
			shouldBeCompressed: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, body := testRequest(t, ts, http.MethodPost, "/", tc.body, tc.headers)
			defer res.Body.Close()

			if tc.shouldBeCompressed {
				assert.Equal(t, tc.expectedBody, decompress(t, body))
			} else {
				assert.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}
