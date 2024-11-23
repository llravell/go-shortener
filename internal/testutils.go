package testutils

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/stretchr/testify/require"
)

const JWTSecretKey = "secret"

func BuildAuthTokenCookie(t *testing.T) *http.Cookie {
	t.Helper()

	jwtToken, err := entity.BuildJWTString("test-uuid", []byte(JWTSecretKey))
	require.NoError(t, err)

	return &http.Cookie{
		Name:  middleware.TokenCookieName,
		Value: jwtToken,
	}
}

func MakeAuthorizedClient(t *testing.T, ts *httptest.Server) *http.Client {
	t.Helper()

	client := ts.Client()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	jar.SetCookies(tsURL, []*http.Cookie{BuildAuthTokenCookie(t)})
	client.Jar = jar

	return client
}
