package service

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	h "github.com/hyphengolang/prelude/http"
)

type Option[T http.Handler] func(T)

type Service interface {
	chi.Router

	Log(...any)
	Logf(string, ...any)

	Decode(http.ResponseWriter, *http.Request, any) error
	Respond(http.ResponseWriter, *http.Request, any, int)
	RespondText(w http.ResponseWriter, r *http.Request, status int)
	Created(http.ResponseWriter, *http.Request, string)
	SetCookie(http.ResponseWriter, *http.Cookie)
}

type serviceHandler struct {
	chi.Router
}

// Created implements Service
func (*serviceHandler) Created(w http.ResponseWriter, r *http.Request, id string) {
	h.Created(w, r, id)
}

// Decode implements Service
func (*serviceHandler) Decode(w http.ResponseWriter, r *http.Request, v any) error {
	return h.Decode(w, r, v)
}

// Log implements Service
func (*serviceHandler) Log(v ...any) { log.Println(v...) }

// Logf implements Service
func (*serviceHandler) Logf(format string, v ...any) { log.Printf(format, v...) }

// Respond implements Service
func (*serviceHandler) Respond(w http.ResponseWriter, r *http.Request, v any, status int) {
	h.Respond(w, r, v, status)
}

func (s *serviceHandler) RespondText(w http.ResponseWriter, r *http.Request, status int) {
	s.Respond(w, r, http.StatusText(status), status)
}

// SetCookie implements Service
func (*serviceHandler) SetCookie(w http.ResponseWriter, c *http.Cookie) { http.SetCookie(w, c) }

func New() Service {
	return &serviceHandler{chi.NewMux()}
}
