package middleware_test

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func findAuthTokenCookie(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	t.Helper()

	var tokenCookie *http.Cookie

	for _, cookie := range cookies {
		if cookie.Name == middleware.TokenCookieName {
			tokenCookie = cookie

			break
		}
	}

	return tokenCookie
}

//nolint:funlen
func TestAuthMiddleware(t *testing.T) {
	router := chi.NewRouter()

	router.Use(middleware.AuthMiddleware("secret"))
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	url, err := url.Parse(ts.URL)
	require.NoError(t, err)

	t.Run("Middleware set auth token cookie", func(t *testing.T) {
		res, _ := testRequest(t, ts, http.MethodPost, "/", http.NoBody, map[string]string{})
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		tokenCookie := findAuthTokenCookie(t, res.Cookies())
		assert.NotEmpty(t, tokenCookie.Value)
	})

	t.Run("Middleware does not change valid token cookie", func(t *testing.T) {
		client := ts.Client()
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)

		client.Jar = jar

		req1, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		_, err = client.Do(req1)
		require.NoError(t, err)

		authToken := findAuthTokenCookie(t, jar.Cookies(url)).Value
		assert.NotEmpty(t, authToken)

		req2, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		res, err := client.Do(req2)
		require.NoError(t, err)

		assert.Nil(t, findAuthTokenCookie(t, res.Cookies()))
		assert.Equal(t, authToken, findAuthTokenCookie(t, jar.Cookies(url)).Value)
	})

	t.Run("Middleware change token cookie if it invalid", func(t *testing.T) {
		client := ts.Client()
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)

		client.Jar = jar

		req1, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		_, err = client.Do(req1)
		require.NoError(t, err)

		authCookie := findAuthTokenCookie(t, jar.Cookies(url))
		authToken := authCookie.Value
		assert.NotEmpty(t, authToken)

		authCookie.Value = "blabla"
		jar.SetCookies(url, []*http.Cookie{authCookie})

		req2, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		res, err := client.Do(req2)
		require.NoError(t, err)

		updatedAuthCookie := findAuthTokenCookie(t, res.Cookies())
		assert.NotEmpty(t, updatedAuthCookie.Value)
		assert.NotEqual(t, authToken, updatedAuthCookie.Value)
	})
}
