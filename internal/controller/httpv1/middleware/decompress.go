package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type decompressor struct{}

func (d decompressor) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "gzip decoding error", http.StatusInternalServerError)
				return
			}

			r.Body = gz
			defer gz.Close()
		}

		next.ServeHTTP(w, r)
	})
}

func DecompressMiddleware() func(next http.Handler) http.Handler {
	d := decompressor{}
	return d.Handler
}
