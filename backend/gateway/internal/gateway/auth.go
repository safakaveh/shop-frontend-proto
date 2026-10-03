package gateway

import (
	"crypto/subtle"
	"net/http"
)

func (g *Gateway) requireServiceToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := []byte("Bearer " + g.cfg.ServiceToken)
		actual := []byte(r.Header.Get("Authorization"))

		if len(expected) != len(actual) ||
			subtle.ConstantTimeCompare(expected, actual) != 1 {
			writeProblem(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r)
	})
}
