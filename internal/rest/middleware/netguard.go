package middleware

import (
	"net"
	"net/http"
)

type netGuard struct {
	subnet *net.IPNet
}

// Handler обработчик мидлвары.
func (ng netGuard) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawIP := r.Header.Get("X-Real-IP")
		if len(rawIP) == 0 {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		ip := net.ParseIP(rawIP)
		if ip == nil {
			w.WriteHeader(http.StatusForbidden)

			return
		}

		if ok := ng.subnet.Contains(ip); ok {
			next.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	})
}

// NetGuardMiddleware проверяет что IP-адрес клиента входит в доверенную подсеть.
func NetGuardMiddleware(subnet *net.IPNet) func(next http.Handler) http.Handler {
	ng := netGuard{
		subnet: subnet,
	}

	return ng.Handler
}
