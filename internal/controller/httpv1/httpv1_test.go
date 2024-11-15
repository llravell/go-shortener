package httpv1_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	name         string
	method       string
	path         string
	prepareMocks func()
	body         io.Reader
	expectedCode int
	expectedBody string
}

func sendTestRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
	body io.Reader,
) (*http.Response, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, ts.URL+path, body)
	require.NoError(t, err)

	res, err := ts.Client().Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(b)
}
