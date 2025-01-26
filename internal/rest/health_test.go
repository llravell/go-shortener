package rest_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/mocks"
	"github.com/llravell/go-shortener/internal/rest"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

var errNoConnection = errors.New("no connection")

func TestHealthRoutes(t *testing.T) {
	repo := mocks.NewMockHealthRepo(gomock.NewController(t))

	healthUseCase := usecase.NewHealthUseCase(repo)
	router := chi.NewRouter()
	logger := zerolog.Nop()
	healthRoutes := rest.NewHealthRoutes(healthUseCase, &logger)

	healthRoutes.Apply(router)

	ts := httptest.NewServer(router)
	defer ts.Close()

	testCases := []testCase{
		{
			name:   "ping with db connection",
			method: http.MethodGet,
			path:   "/ping",
			prepareMocks: func() {
				repo.EXPECT().
					PingContext(gomock.Any()).
					Return(nil)
			},
			expectedCode: http.StatusOK,
		},
		{
			name:   "ping without db connection",
			method: http.MethodGet,
			path:   "/ping",
			prepareMocks: func() {
				repo.EXPECT().
					PingContext(gomock.Any()).
					Return(errNoConnection)
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
