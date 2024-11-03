package httpv1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	Hash            = "ABC"
	URL             = "https://example.ru/"
	BaseRedirectURL = "http://localhost:8080"
)

var redirectURL = fmt.Sprintf("%s/%s", BaseRedirectURL, Hash)

var errNotFound = errors.New("not found")

type MockHashGenerator struct{}

func (g MockHashGenerator) Generate() (string, error) {
	return Hash, nil
}

type MockRepo struct {
	m map[string]*entity.URL
}

func (g *MockRepo) Store(_ *entity.URL)    {}
func (g *MockRepo) GetList() []*entity.URL { return []*entity.URL{} }
func (g *MockRepo) Get(hash string) (*entity.URL, error) {
	v, ok := g.m[hash]
	if !ok {
		return nil, errNotFound
	}

	return v, nil
}

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method string,
	path string,
	body io.Reader,
) (*http.Response, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.TODO(), method, ts.URL+path, body)
	require.NoError(t, err)

	res, err := ts.Client().Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, string(b)
}

//nolint:funlen
func TestURL(t *testing.T) {
	gen := MockHashGenerator{}
	s := &MockRepo{map[string]*entity.URL{
		Hash: entity.NewURL(redirectURL, Hash),
	}}

	urlUseCase := usecase.NewURLUseCase(s, gen, BaseRedirectURL)
	router := chi.NewRouter()
	logger := zerolog.Nop()
	newURLRoutes(router, urlUseCase, logger)

	ts := httptest.NewServer(router)
	ts.Client().CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
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
			body:         strings.NewReader(fmt.Sprintf(`{"url":"%s"}`, URL)),
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
			res, body := testRequest(t, ts, tc.method, tc.path, tc.body)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)

			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, body)
			}
		})
	}
}
