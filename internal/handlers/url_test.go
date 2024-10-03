package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/llravell/go-shortener/internal/storages"
	"github.com/stretchr/testify/assert"
)

const HASH = "ABC"

type MockHashGenerator struct{}

func (g MockHashGenerator) Generate(n int) string {
	return HASH
}

func TestSaveUrlHandler(t *testing.T) {
	us := storages.NewUrlStorage()
	gen := MockHashGenerator{}

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
