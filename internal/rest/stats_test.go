package rest_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/mocks"
	"github.com/llravell/go-shortener/internal/rest"
	"github.com/llravell/go-shortener/internal/usecase"
)

var errFetchingStats = errors.New("fetching stats failed")

//nolint:funlen
func TestStatsRoutes(t *testing.T) {
	repo := mocks.NewMockStatsRepo(gomock.NewController(t))

	statsUseCase := usecase.NewStatsUseCase(repo)
	router := chi.NewRouter()
	logger := zerolog.Nop()
	statsRoutes := rest.NewStatsRoutes(statsUseCase, &logger)

	statsRoutes.Apply(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	testCases := []testCase{
		{
			name:   "return app stats",
			method: http.MethodGet,
			path:   "/stats",
			prepareMocks: func() {
				repo.EXPECT().
					GetURLsAmount(gomock.Any()).
					Return(10, nil)

				repo.EXPECT().
					GetUsersAmount(gomock.Any()).
					Return(5, nil)
			},
			expectedCode: http.StatusOK,
			expectedBody: toJSON(t, map[string]int{
				"urls":  10,
				"users": 5,
			}),
		},
		{
			name:   "urls stats fetching failed",
			method: http.MethodGet,
			path:   "/stats",
			prepareMocks: func() {
				repo.EXPECT().
					GetURLsAmount(gomock.Any()).
					Return(0, errFetchingStats)

				repo.EXPECT().
					GetUsersAmount(gomock.Any()).
					Return(5, nil)
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:   "users stats fetching failed",
			method: http.MethodGet,
			path:   "/stats",
			prepareMocks: func() {
				repo.EXPECT().
					GetURLsAmount(gomock.Any()).
					Return(10, nil)

				repo.EXPECT().
					GetUsersAmount(gomock.Any()).
					Return(0, errFetchingStats)
			},
			expectedCode: http.StatusInternalServerError,
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
