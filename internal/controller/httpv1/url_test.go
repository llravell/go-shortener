package httpv1

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const Hash = "ABC"
const URL = "https://example.ru/"
const BaseRedirectURL = "http://localhost:8080"

var redirectURL = fmt.Sprintf("%s/%s", BaseRedirectURL, Hash)

type MockHashGenerator struct{}

func (g MockHashGenerator) Generate(url string) string {
	return Hash
}

type MockRepo struct {
	m map[string]string
}

func (g *MockRepo) Store(string, string) {}
func (g *MockRepo) Get(hash string) (string, error) {
	v, ok := g.m[hash]
	if !ok {
		return "", errors.New("Not found")
	}

	return v, nil
}

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

func TestURL(t *testing.T) {
	gen := MockHashGenerator{}
	s := &MockRepo{map[string]string{
		Hash: URL,
	}}

	urlUseCase := usecase.NewURLUseCase(s, gen)
	router := chi.NewRouter()
	logger := zerolog.Nop()
	newURLRoutes(router, urlUseCase, logger, BaseRedirectURL)

	ts := httptest.NewServer(router)
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	defer ts.Close()

	defaultBody := fmt.Sprintf(`{"url":"%s"}`, URL)

	compressedBody := &bytes.Buffer{}
	wr := gzip.NewWriter(compressedBody)
	wr.Write([]byte(defaultBody))
	wr.Close()

	testCases := []struct {
		name         string
		method       string
		path         string
		body         io.Reader
		headers      map[string]string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "[legacy] sending url",
			method:       http.MethodPost,
			path:         "/",
			body:         strings.NewReader(URL),
			expectedCode: http.StatusCreated,
			expectedBody: fmt.Sprintf("%s/%s", BaseRedirectURL, Hash),
		},
		{
			name:         "[legacy] sending empty payload",
			method:       http.MethodPost,
			path:         "/",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Sending url",
			method:       http.MethodPost,
			path:         "/api/shorten",
			body:         strings.NewReader(defaultBody),
			expectedCode: http.StatusCreated,
			expectedBody: fmt.Sprintf("{\"result\":\"%s\"}\n", redirectURL),
		},
		{
			name:   "Sending url with gzip compression",
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   compressedBody,
			headers: map[string]string{
				"Content-Encoding": "gzip",
			},
			expectedCode: http.StatusCreated,
			expectedBody: fmt.Sprintf("{\"result\":\"%s\"}\n", redirectURL),
		},
		{
			name:         "Sending empty payload",
			method:       http.MethodPost,
			path:         "/api/shorten",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Redirect on url",
			method:       http.MethodGet,
			path:         "/" + Hash,
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
			res, body := testRequest(t, ts, tc.method, tc.path, tc.body, tc.headers)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, body)
			}
		})
	}
}
