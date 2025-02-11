package middleware

import (
	"compress/flate"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
)

// CompressMiddleware мидлвара сжатия данных http ответа.
func CompressMiddleware(types ...string) func(next http.Handler) http.Handler {
	return middleware.Compress(
		flate.DefaultCompression,
		types...,
	)
}
