package service

import (
	"bytes"
	"encoding/json"
	"net/http"

	srv "github.com/rog-golang-buddies/rmx/internal/http"
	"github.com/rog-golang-buddies/rmx/internal/sql"
	"github.com/rog-golang-buddies/rmx/internal/user"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
)

type userService struct {
	mux srv.Service
	// Base URL for auth service
	authSrvURL string
	// Repo for users
	r sql.RWRepo[user.User]
}

func (s *userService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func NewService(opts ...UserOption) http.Handler {
	s := &userService{
		mux: srv.New(),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.r == nil {
		panic("user repo is required")
	}

	if s.authSrvURL == "" {
		panic("auth service url is required")
	}

	s.routes()
	return s
}

func (s *userService) routes() {
	s.mux.Post("/api/v1/user/register", s.handleRegister())
	s.mux.Delete("/api/v1/user/register", s.handleUnregister())
	// NOTE - debug only
	s.mux.Get("/api/v1/user/ping", s.handlePing())
}

func (s *userService) handleRegister() http.HandlerFunc {
	type User struct {
		Email    email.Email       `json:"email"`
		Username string            `json:"username"`
		Password password.Password `json:"password"`
	}

	type credentials struct {
		Email    email.Email       `json:"email"`
		Password password.Password `json:"password"`
	}

	newUser := func(w http.ResponseWriter, r *http.Request, u *user.User) (err error) {
		var dto User
		if err = s.mux.Decode(w, r, &dto); err != nil {
			return
		}

		*u = user.User{
			ID:       suid.NewUUID(),
			Username: dto.Username,
			Email:    dto.Email,
			// NOTE this could be a gob encoded message
			// and so won't need worry about this being
			// visible in the request body.
			Password: dto.Password,
		}

		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var u user.User
		if err := newUser(w, r, &u); err != nil {
			s.mux.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.r.Write(r.Context(), &u); err != nil {
			s.mux.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		{
			// make an api call to the auth service to store user
			// credentials. this can be achieved by making a POST
			// request to the auth service's /api/v1/auth/register
			// endpoint.

			cred := &credentials{
				Email:    u.Email,
				Password: u.Password,
			}

			var p []byte
			var err error
			if p, err = json.Marshal(cred); err != nil {
				s.mux.Respond(w, r, err, http.StatusInternalServerError)
				return
			}

			resp, err := http.Post(s.authSrvURL+"/api/v1/auth/credentials", "application/json", bytes.NewReader(p))
			if err != nil {
				s.mux.Respond(w, r, err, http.StatusInternalServerError)
				return
			}

			if resp.StatusCode != http.StatusCreated {
				// NOTE - new line
				err := s.r.Remove(r.Context(), u.ID)
				s.mux.Respond(w, r, err, http.StatusInternalServerError)
				return
			}
		}

		suid := u.ID.ShortUUID().String()
		s.mux.Created(w, r, suid)
	}
}

func (s *userService) handleUnregister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.Respond(w, r, nil, http.StatusNotImplemented)
	}
}

func (s *userService) handlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.Respond(w, r, nil, http.StatusNoContent)
	}
}

type UserOption srv.Option[*userService]

func WithRepo(r sql.RWRepo[user.User]) UserOption {
	return func(s *userService) {
		s.r = r
	}
}

func WithAuthServiceURL(url string) UserOption {
	return func(s *userService) {
		s.authSrvURL = url
	}
}
