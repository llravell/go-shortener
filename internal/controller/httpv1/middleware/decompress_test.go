package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader, headers map[string]string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := ts.Client().Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(b)
}

func compressReader(t *testing.T, data string) io.Reader {
	var buf bytes.Buffer

	wr := gzip.NewWriter(&buf)
	_, err := wr.Write([]byte(data))
	require.NoError(t, err)

	err = wr.Close()
	require.NoError(t, err)

	return &buf
}

func TestDecompressMiddleware(t *testing.T) {
	router := chi.NewRouter()

	router.Use(DecompressMiddleware())
	router.Post("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write(body)
	})

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

			assert.Equal(t, tc.expectedBody, body)
		})
	}
}
