package www

import "net/http"

func (s Service) routes() {
	fs := http.FileServer(http.Dir("assets"))
	s.m.Handle("/assets/*", http.StripPrefix("/assets/", fs))

	// s.m.Handle("/assets/*", s.fileServer("/assets/", "assets"))

	s.m.Get("/ping", s.handlePing())
}

func (s Service) handlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, "pong", http.StatusOK)
	}
}
