package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
	body io.Reader,
	headers map[string]string,
) (*http.Response, []byte) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := ts.Client().Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, b
}

func echoHandler(t *testing.T) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		w.Header().Set("Content-Type", "text/plain")

		_, err = w.Write(body)
		require.NoError(t, err)
	})
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

func decompress(t *testing.T, data []byte) string {
	buf := bytes.NewBuffer(data)
	r, err := gzip.NewReader(buf)
	require.NoError(t, err)

	res, err := io.ReadAll(r)
	require.NoError(t, err)

	err = r.Close()
	require.NoError(t, err)

	return string(res)
}
