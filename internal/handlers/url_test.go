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

const HASH = "ABC"
const URL = "https://example.ru/"

type MockHashGenerator struct{}

func (g MockHashGenerator) Generate(n int) string {
	return HASH
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

func TestSaveUrlHandler(t *testing.T) {
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
			expectedBody: fmt.Sprintf("http://localhost:8080/%s", HASH),
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
			us := storages.NewUrlStorage()
			gen := MockHashGenerator{}

			handler := SaveUrlHandler(us, gen)
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

func TestResolveUrlHandler(t *testing.T) {
	testCases := []struct {
		name          string
		hash          string
		storageResult string
		expectedCode  int
	}{
		{
			name:          "Redirect on url",
			hash:          HASH,
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

			handler := ResolveUrlHandler(us)
			r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/%s", tc.hash), nil)
			w := httptest.NewRecorder()

			handler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}
