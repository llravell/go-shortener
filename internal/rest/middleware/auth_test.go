package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-chi/chi/v5"
	testutils "github.com/llravell/go-shortener/internal"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/rs/zerolog"
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

func TestProvideJWTMiddleware(t *testing.T) {
	router := chi.NewRouter()
	logger := zerolog.Nop()
	auth := middleware.NewAuth(testutils.JWTSecretKey, &logger)

	router.Use(auth.ProvideJWTMiddleware)
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	t.Run("Middleware set auth token cookie", func(t *testing.T) {
		res, _ := testutils.SendTestRequest(t, ts, ts.Client(), http.MethodPost, "/", http.NoBody, map[string]string{})
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		tokenCookie := findAuthTokenCookie(t, res.Cookies())
		assert.NotEmpty(t, tokenCookie.Value)
	})

	t.Run("Middleware does not change valid token cookie", func(t *testing.T) {
		client := testutils.AuthorizedClient(t, ts)
		authToken := findAuthTokenCookie(t, client.Jar.Cookies(tsURL)).Value

		res, _ := testutils.SendTestRequest(t, ts, client, http.MethodPost, "/", http.NoBody, map[string]string{})
		defer res.Body.Close()

		assert.Nil(t, findAuthTokenCookie(t, res.Cookies()))
		assert.Equal(t, authToken, findAuthTokenCookie(t, client.Jar.Cookies(tsURL)).Value)
	})

	t.Run("Middleware change token cookie if it invalid", func(t *testing.T) {
		client := testutils.AuthorizedClient(t, ts)
		authCookie := findAuthTokenCookie(t, client.Jar.Cookies(tsURL))
		authToken := authCookie.Value
		assert.NotEmpty(t, authToken)

		authCookie.Value = "blabla"
		client.Jar.SetCookies(tsURL, []*http.Cookie{authCookie})

		res, _ := testutils.SendTestRequest(t, ts, client, http.MethodPost, "/", http.NoBody, map[string]string{})
		defer res.Body.Close()

		updatedAuthCookie := findAuthTokenCookie(t, res.Cookies())
		assert.NotEmpty(t, updatedAuthCookie.Value)
		assert.NotEqual(t, authToken, updatedAuthCookie.Value)
	})
}

func TestCheckJWTMiddleware(t *testing.T) {
	router := chi.NewRouter()
	logger := zerolog.Nop()
	auth := middleware.NewAuth(testutils.JWTSecretKey, &logger)

	router.Use(auth.CheckJWTMiddleware)
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	t.Run("Middleware return unauthorized status code if token does not exist", func(t *testing.T) {
		res, _ := testutils.SendTestRequest(t, ts, ts.Client(), http.MethodPost, "/", http.NoBody, map[string]string{})
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("Middleware call original handler if token exists", func(t *testing.T) {
		res, _ := testutils.SendTestRequest(
			t, ts, testutils.AuthorizedClient(t, ts), http.MethodPost, "/", http.NoBody, map[string]string{},
		)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
