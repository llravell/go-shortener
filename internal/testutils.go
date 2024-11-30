package testutils

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/stretchr/testify/require"
)

const JWTSecretKey = "secret"

func buildAuthTokenCookie(t *testing.T) *http.Cookie {
	t.Helper()

	jwtToken, err := entity.BuildJWTString("test-uuid", []byte(JWTSecretKey))
	require.NoError(t, err)

	return &http.Cookie{
		Name:  middleware.TokenCookieName,
		Value: jwtToken,
	}
}

func AuthorizedClient(t *testing.T, ts *httptest.Server) *http.Client {
	t.Helper()

	client := *ts.Client()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	tsURL, err := url.Parse(ts.URL)
	require.NoError(t, err)

	jar.SetCookies(tsURL, []*http.Cookie{buildAuthTokenCookie(t)})
	client.Jar = jar

	return &client
}

func SendTestRequest(
	t *testing.T,
	ts *httptest.Server,
	client *http.Client,
	method string,
	path string,
	body io.Reader,
	headers map[string]string,
) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequestWithContext(context.TODO(), method, ts.URL+path, body)
	require.NoError(t, err)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := client.Do(req)
	require.NoError(t, err)

	b, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	return res, b
}
