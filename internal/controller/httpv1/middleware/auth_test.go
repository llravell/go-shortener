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
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const jwtSecretKey = "secret"

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

func buildAuthTokenCookie(t *testing.T) *http.Cookie {
	t.Helper()

	jwtToken, err := entity.BuildJWTString("test-uuid", []byte(jwtSecretKey))
	require.NoError(t, err)

	return &http.Cookie{
		Name:  middleware.TokenCookieName,
		Value: jwtToken,
	}
}

//nolint:funlen
func TestProvideJWTMiddleware(t *testing.T) {
	router := chi.NewRouter()
	auth := middleware.NewAuth(jwtSecretKey)

	router.Use(auth.ProvideJWTMiddleware)
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	tsURL, err := url.Parse(ts.URL)
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

		req1, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		_, err = client.Do(req1)
		require.NoError(t, err)

		authToken := findAuthTokenCookie(t, jar.Cookies(tsURL)).Value
		assert.NotEmpty(t, authToken)

		req2, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		res, err := client.Do(req2)
		require.NoError(t, err)

		assert.Nil(t, findAuthTokenCookie(t, res.Cookies()))
		assert.Equal(t, authToken, findAuthTokenCookie(t, jar.Cookies(tsURL)).Value)
	})

	t.Run("Middleware change token cookie if it invalid", func(t *testing.T) {
		client := ts.Client()
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)

		client.Jar = jar

		req1, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		_, err = client.Do(req1)
		require.NoError(t, err)

		authCookie := findAuthTokenCookie(t, jar.Cookies(tsURL))
		authToken := authCookie.Value
		assert.NotEmpty(t, authToken)

		authCookie.Value = "blabla"
		jar.SetCookies(tsURL, []*http.Cookie{authCookie})

		req2, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		res, err := client.Do(req2)
		require.NoError(t, err)

		updatedAuthCookie := findAuthTokenCookie(t, res.Cookies())
		assert.NotEmpty(t, updatedAuthCookie.Value)
		assert.NotEqual(t, authToken, updatedAuthCookie.Value)
	})
}

func TestCheckJWTMiddleware(t *testing.T) {
	router := chi.NewRouter()
	auth := middleware.NewAuth(jwtSecretKey)

	router.Use(auth.CheckJWTMiddleware)
	router.Post("/", echoHandler(t))

	ts := httptest.NewServer(router)

	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	t.Run("Middleware return unauthorized status code if token does not exist", func(t *testing.T) {
		res, _ := testRequest(t, ts, http.MethodPost, "/", http.NoBody, map[string]string{})
		defer res.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("Middleware call original handler if token exists", func(t *testing.T) {
		client := ts.Client()
		jar, err := cookiejar.New(nil)
		require.NoError(t, err)

		jar.SetCookies(tsURL, []*http.Cookie{buildAuthTokenCookie(t)})
		client.Jar = jar

		req, err := http.NewRequestWithContext(context.TODO(), http.MethodPost, ts.URL+"/", http.NoBody)
		require.NoError(t, err)

		res, err := client.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}
