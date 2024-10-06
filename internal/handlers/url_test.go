package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/llravell/go-shortener/internal/storages"
	"github.com/stretchr/testify/assert"
)

const Hash = "ABC"
const URL = "https://example.ru/"

type MockHashGenerator struct{}

func (g MockHashGenerator) Generate(n int) string {
	return Hash
}

type MockStorage struct {
	result string
}

func (g *MockStorage) Save(string, string) {}
func (g *MockStorage) Get(string) (string, error) {
	if g.result != "" {
		return g.result, nil
	}

	return "", errors.New("Error")
}

func TestSaveURLHandler(t *testing.T) {
	testCases := []struct {
		name         string
		payload      string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Sending url",
			payload:      "https://example.ru/",
			expectedCode: http.StatusCreated,
			expectedBody: fmt.Sprintf("http://localhost:8080/%s", Hash),
		},
		{
			name:         "Sending empty payload",
			payload:      "",
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			us := storages.NewURLStorage()
			gen := MockHashGenerator{}

			handler := SaveURLHandler(us, gen)
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.payload))
			w := httptest.NewRecorder()

			handler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String())
			}
		})
	}
}

func TestResolveURLHandler(t *testing.T) {
	testCases := []struct {
		name          string
		hash          string
		storageResult string
		expectedCode  int
	}{
		{
			name:          "Redirect on url",
			hash:          Hash,
			storageResult: URL,
			expectedCode:  http.StatusTemporaryRedirect,
		},
		{
			name:          "Failed redirect",
			hash:          "",
			storageResult: "",
			expectedCode:  http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			us := &MockStorage{tc.storageResult}

			handler := ResolveURLHandler(us)
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tc.hash), nil)
			w := httptest.NewRecorder()

			handler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}
