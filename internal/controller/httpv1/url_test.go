package httpv1

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/mocks"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errNotFound = errors.New("not found")

type testCase struct {
	name         string
	method       string
	path         string
	prepareMocks func()
	body         io.Reader
	expectedCode int
	expectedBody string
}

func toJSON(t *testing.T, m any) string {
	t.Helper()

	data, err := json.Marshal(m)
	require.NoError(t, err)

	data = append(data, '\n')

	return string(data)
}

func testRequest(
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

func prepareTestServer(gen usecase.HashGenerator, repo usecase.URLRepo) *httptest.Server {
	urlUseCase := usecase.NewURLUseCase(repo, gen, "http://localhost:8080")
	router := chi.NewRouter()
	logger := zerolog.Nop()
	newURLRoutes(router, urlUseCase, logger)

	ts := httptest.NewServer(router)
	ts.Client().CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return ts
}

//nolint:funlen
func TestURLBaseRoutes(t *testing.T) {
	gen := mocks.NewMockHashGenerator(gomock.NewController(t))
	repo := mocks.NewMockURLRepo(gomock.NewController(t))

	gen.EXPECT().Generate().AnyTimes()

	ts := prepareTestServer(gen, repo)
	defer ts.Close()

	testCases := []testCase{
		{
			name:   "[legacy] sending url",
			method: http.MethodPost,
			path:   "/",
			prepareMocks: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(&entity.URL{Short: "a"}, nil)
			},
			body:         strings.NewReader("https://a.ru"),
			expectedCode: http.StatusCreated,
			expectedBody: "http://localhost:8080/a",
		},
		{
			name:         "[legacy] sending empty payload",
			method:       http.MethodPost,
			path:         "/",
			prepareMocks: func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "Sending url",
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   strings.NewReader(toJSON(t, map[string]string{"url": "https://a.ru"})),
			prepareMocks: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(&entity.URL{Short: "a"}, nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: toJSON(t, map[string]string{"result": "http://localhost:8080/a"}),
		},
		{
			name:         "Sending empty payload",
			method:       http.MethodPost,
			path:         "/api/shorten",
			prepareMocks: func() {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:   "Redirect on url",
			method: http.MethodGet,
			path:   "/a",
			prepareMocks: func() {
				repo.EXPECT().
					Get(gomock.Any(), "a").
					Return(&entity.URL{Original: "https://a.ru"}, nil)
			},
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:   "Failed redirect",
			method: http.MethodGet,
			path:   "/not_existed_hash",
			prepareMocks: func() {
				repo.EXPECT().
					Get(gomock.Any(), "not_existed_hash").
					Return(nil, errNotFound)
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.prepareMocks()

			res, body := testRequest(t, ts, tc.method, tc.path, tc.body)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)

			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, body)
			}
		})
	}
}

//nolint:funlen
func TestURLBatchRoute(t *testing.T) {
	gen := mocks.NewMockHashGenerator(gomock.NewController(t))
	repo := mocks.NewMockURLRepo(gomock.NewController(t))

	ts := prepareTestServer(gen, repo)
	defer ts.Close()

	testCases := []testCase{
		{
			name:   "Sending several urls",
			method: http.MethodPost,
			path:   "/api/shorten/batch",
			body: strings.NewReader(toJSON(t, []map[string]string{
				{
					"correlation_id": "1",
					"original_url":   "https://a.ru",
				},
				{
					"correlation_id": "2",
					"original_url":   "https://b.ru",
				},
			})),
			prepareMocks: func() {
				gomock.InOrder(
					gen.EXPECT().Generate().Return("a", nil),
					gen.EXPECT().Generate().Return("b", nil),
				)

				repo.EXPECT().
					StoreMultiple(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusCreated,
			expectedBody: toJSON(t, []map[string]string{
				{
					"correlation_id": "1",
					"short_url":      "http://localhost:8080/a",
				},
				{
					"correlation_id": "2",
					"short_url":      "http://localhost:8080/b",
				},
			}),
		},
		{
			name:         "Sending empty urls",
			method:       http.MethodPost,
			path:         "/api/shorten/batch",
			body:         strings.NewReader(toJSON(t, []any{})),
			prepareMocks: func() {},
			expectedCode: http.StatusCreated,
			expectedBody: toJSON(t, []any{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.prepareMocks()

			res, body := testRequest(t, ts, tc.method, tc.path, tc.body)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)

			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, body)
			}
		})
	}
}
