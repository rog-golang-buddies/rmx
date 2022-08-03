package www

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Service struct {
	m *chi.Mux
}

func NewService() *Service {
	s := &Service{
		m: chi.NewMux(),
	}
	s.routes()
	return s
}

func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.m.ServeHTTP(w, r)
}

func (s Service) respond(rw http.ResponseWriter, r *http.Request, data interface{}, status int) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(status)
	if data != nil {
		err := json.NewEncoder(rw).Encode(data)
		if err != nil {
			http.Error(rw, "Could not encode in json", status)
		}
	}
}

func (s Service) decode(rw http.ResponseWriter, r *http.Request, data interface{}) (err error) {
	return json.NewDecoder(r.Body).Decode(data)
}

func (s Service) created(rw http.ResponseWriter, r *http.Request, id string) {
	rw.Header().Add("Location", "//"+r.Host+r.URL.Path+"/"+id)
	s.respond(rw, r, nil, http.StatusCreated)
}

func (s Service) fileServer(prefix, dirname string) http.Handler {
	return http.StripPrefix(prefix, http.FileServer(http.Dir(dirname)))
}
