package req

import (
	"golang.org/x/exp/slices"
	"net/http"
)

func CheckMethod(m []string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if ok := slices.Contains(m, r.Method); !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
