package httpv1_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	testutils "github.com/llravell/go-shortener/internal"
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
	withAuth bool,
) (*http.Response, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.Background(), method, ts.URL+path, body)
	require.NoError(t, err)

	client := ts.Client()

	if withAuth {
		client = testutils.MakeAuthorizedClient(t, ts)
	}

	res, err := client.Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(b)
}
