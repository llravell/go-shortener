package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const Hash = "ABC"
const URL = "https://example.ru/"
const BaseRedirectURL = "http://localhost:8080"

type MockHashGenerator struct{}

func (g MockHashGenerator) Generate(n int) string {
	return Hash
}

type MockStorage struct {
	m map[string]string
}

func (g *MockStorage) Save(string, string) {}
func (g *MockStorage) Get(hash string) (string, error) {
	v, ok := g.m[hash]
	if !ok {
		return "", errors.New("Not found")
	}

	return v, nil
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	res, err := ts.Client().Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(b)
}

func TestBuildRouter(t *testing.T) {
	gen := MockHashGenerator{}
	s := &MockStorage{map[string]string{
		Hash: URL,
	}}

	ts := httptest.NewServer(BuildRouter(s, gen, BaseRedirectURL))
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer ts.Close()

	testCases := []struct {
		name         string
		method       string
		path         string
		body         io.Reader
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Sending url",
			method:       http.MethodPost,
			path:         "/",
			body:         strings.NewReader(URL),
			expectedCode: http.StatusCreated,
			expectedBody: fmt.Sprintf("%s/%s", BaseRedirectURL, Hash),
		},
		{
			name:         "Sending empty payload",
			method:       http.MethodPost,
			path:         "/",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Redirect on url",
			method:       http.MethodGet,
			path:         "/" + Hash,
			body:         nil,
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "Failed redirect",
			method:       http.MethodGet,
			path:         "/not_existed_hash",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, body := testRequest(t, ts, tc.method, tc.path, tc.body)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, body)
			}
		})
	}
}
