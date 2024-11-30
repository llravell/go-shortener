package rest_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/mocks"
	repository "github.com/llravell/go-shortener/internal/repo"
	"github.com/llravell/go-shortener/internal/rest"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errNotFound = errors.New("not found")

type urlDeleteWorkMatcher struct {
	userUUID string
	hashes   []string
}

func (m *urlDeleteWorkMatcher) Matches(x interface{}) bool {
	urlDeleteWork, ok := x.(*usecase.URLDeleteWork)
	if !ok {
		return false
	}

	return (urlDeleteWork.UserUUID == m.userUUID &&
		reflect.DeepEqual(urlDeleteWork.Hashes, m.hashes))
}

func (m *urlDeleteWorkMatcher) String() string {
	return fmt.Sprintf("has userUUID=\"%s\" and hashes=\"%v\"", m.userUUID, m.hashes)
}

func toJSON(t *testing.T, m any) string {
	t.Helper()

	data, err := json.Marshal(m)
	require.NoError(t, err)

	data = append(data, '\n')

	return string(data)
}

func prepareTestServer(
	gen usecase.HashGenerator,
	repo usecase.URLRepo,
	wp usecase.URLDeleteWorkerPool,
) *httptest.Server {
	logger := zerolog.Nop()

	urlUseCase := usecase.NewURLUseCase(repo, gen, "http://localhost:8080")
	urlDeleteUseCase := usecase.NewURLDeleteUseCase(repo, wp, logger)

	router := chi.NewRouter()
	auth := middleware.NewAuth("secret", logger)
	rest.NewURLRoutes(router, urlUseCase, urlDeleteUseCase, auth, logger)

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
	wp := mocks.NewMockURLDeleteWorkerPool(gomock.NewController(t))

	gen.EXPECT().Generate().AnyTimes()

	ts := prepareTestServer(gen, repo, wp)
	defer ts.Close()

	testCases := []testCase{
		{
			name:   "[legacy] sending url",
			method: http.MethodPost,
			path:   "/",
			prepareMocks: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(&entity.URL{Short: "a"}, repository.ErrOriginalURLConflict)
			},
			body:         strings.NewReader("https://a.ru"),
			expectedCode: http.StatusConflict,
			expectedBody: "http://localhost:8080/a",
		},
		{
			name:   "[legacy] sending already existed url",
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
			name:   "Sending already existed url",
			method: http.MethodPost,
			path:   "/api/shorten",
			body:   strings.NewReader(toJSON(t, map[string]string{"url": "https://a.ru"})),
			prepareMocks: func() {
				repo.EXPECT().
					Store(gomock.Any(), gomock.Any()).
					Return(&entity.URL{Short: "a"}, repository.ErrOriginalURLConflict)
			},
			expectedCode: http.StatusConflict,
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
		{
			name:   "Redirect on deleted url",
			method: http.MethodGet,
			path:   "/deleted_url",
			prepareMocks: func() {
				repo.EXPECT().
					Get(gomock.Any(), "deleted_url").
					Return(&entity.URL{Deleted: true}, nil)
			},
			expectedCode: http.StatusGone,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.prepareMocks()

			res, body := testutils.SendTestRequest(t, ts, ts.Client(), tc.method, tc.path, tc.body, map[string]string{})
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)

			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}

//nolint:funlen
func TestURLBatchRoute(t *testing.T) {
	gen := mocks.NewMockHashGenerator(gomock.NewController(t))
	repo := mocks.NewMockURLRepo(gomock.NewController(t))
	wp := mocks.NewMockURLDeleteWorkerPool(gomock.NewController(t))

	ts := prepareTestServer(gen, repo, wp)
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

			res, body := testutils.SendTestRequest(t, ts, ts.Client(), tc.method, tc.path, tc.body, map[string]string{})
			defer res.Body.Close()

			assert.Equal(t, tc.expectedCode, res.StatusCode)

			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, string(body))
			}
		})
	}
}

//nolint:funlen
func TestURLUserRoutes(t *testing.T) {
	gen := mocks.NewMockHashGenerator(gomock.NewController(t))
	repo := mocks.NewMockURLRepo(gomock.NewController(t))
	wp := mocks.NewMockURLDeleteWorkerPool(gomock.NewController(t))

	ts := prepareTestServer(gen, repo, wp)
	defer ts.Close()

	t.Run("Return unauthorized status code for unauthorized get try", func(t *testing.T) {
		res, _ := testutils.SendTestRequest(
			t, ts, ts.Client(), http.MethodGet, "/api/user/urls", http.NoBody, map[string]string{},
		)
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("Return no content status code user without urls", func(t *testing.T) {
		repo.EXPECT().
			GetByUserUUID(gomock.Any(), gomock.Any()).
			Return([]*entity.URL{}, nil)

		res, _ := testutils.SendTestRequest(
			t, ts, testutils.AuthorizedClient(t, ts), http.MethodGet, "/api/user/urls", http.NoBody, map[string]string{},
		)
		defer res.Body.Close()

		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})

	t.Run("Return user's urls", func(t *testing.T) {
		repo.EXPECT().
			GetByUserUUID(gomock.Any(), gomock.Any()).
			Return([]*entity.URL{
				{
					Short:    "a",
					Original: "https://a.ru",
				},
			}, nil)

		res, body := testutils.SendTestRequest(
			t, ts, testutils.AuthorizedClient(t, ts), http.MethodGet, "/api/user/urls", http.NoBody, map[string]string{},
		)
		defer res.Body.Close()

		expectedBody := toJSON(t, []rest.UserURLItem{
			{
				ShortURL:    "http://localhost:8080/a",
				OriginalURL: "https://a.ru",
			},
		})

		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Equal(t, expectedBody, string(body))
	})

	t.Run("Return unauthorized status code for unauthorized delete try", func(t *testing.T) {
		res, _ := testutils.SendTestRequest(
			t, ts, ts.Client(), http.MethodDelete, "/api/user/urls", http.NoBody, map[string]string{},
		)
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("Successful deleting several urls", func(t *testing.T) {
		hashes := []string{"a", "b"}

		var workMatcher gomock.Matcher = &urlDeleteWorkMatcher{
			userUUID: testutils.UserUUID,
			hashes:   hashes,
		}

		wp.EXPECT().QueueWork(workMatcher).Return(nil)

		body := strings.NewReader(toJSON(t, hashes))
		res, _ := testutils.SendTestRequest(
			t, ts, testutils.AuthorizedClient(t, ts), http.MethodDelete, "/api/user/urls", body, map[string]string{},
		)

		defer res.Body.Close()

		assert.Equal(t, http.StatusAccepted, res.StatusCode)
	})
}
