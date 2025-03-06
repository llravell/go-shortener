package middleware_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/rest/middleware"
)

func TestNetGuardMiddleware(t *testing.T) {
	router := chi.NewRouter()

	subnet := &net.IPNet{
		IP:   net.IP{192, 168, 1, 0},
		Mask: net.IPMask{255, 255, 255, 0},
	}

	router.Use(middleware.NetGuardMiddleware(subnet))
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	testCases := []struct {
		name           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Try getting access without IP header",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Try getting access with IP that is not part of subnet",
			expectedStatus: http.StatusForbidden,
			headers: map[string]string{
				"X-Real-IP": "192.168.100.0",
			},
		},
		{
			name:           "Got access with IP that is included in subnet",
			expectedStatus: http.StatusOK,
			headers: map[string]string{
				"X-Real-IP": "192.168.1.20",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, _ := testutils.SendTestRequest(t, ts, ts.Client(), http.MethodPost, "/", nil, tc.headers)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatus, res.StatusCode)
		})
	}
}
