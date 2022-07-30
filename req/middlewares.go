package req

import (
	"net/http"

	"golang.org/x/exp/slices"
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
