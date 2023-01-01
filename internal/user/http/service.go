package service

import (
	"net/http"

	srv "github.com/rog-golang-buddies/rmx/internal/http"
	"github.com/rog-golang-buddies/rmx/internal/sql"
	"github.com/rog-golang-buddies/rmx/internal/user"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
)

type UserOption srv.Option[*userService]

func WithRepo(r sql.RWRepo[user.User]) UserOption {
	return func(s *userService) {
		s.r = r
	}
}

type userService struct {
	mux srv.Service
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

	s.routes()
	return s
}

func (s *userService) routes() {
	s.mux.Post("/api/v1/user/register", s.handleRegister())
	s.mux.Delete("/api/v1/user/register", s.handleUnregister())
	// NOTE - debug only
	s.mux.Get("/api/v1/user/ping", s.handleHealth())
}

func (s *userService) handleRegister() http.HandlerFunc {
	type User struct {
		Email    email.Email       `json:"email"`
		Username string            `json:"username"`
		Password password.Password `json:"password"`
	}

	newUser := func(w http.ResponseWriter, r *http.Request, u *user.User) (err error) {
		var dto User
		if err = s.mux.Decode(w, r, &dto); err != nil {
			return
		}

		var h password.PasswordHash
		h, err = dto.Password.Hash()
		if err != nil {
			return
		}

		*u = user.User{
			ID:       suid.NewUUID(),
			Username: dto.Username,
			Email:    dto.Email,
			Password: h,
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

		suid := u.ID.ShortUUID().String()

		{
			// make an api call to the auth service to store user
			// credentials
		}

		s.mux.Created(w, r, suid)
	}
}

func (s *userService) handleUnregister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.Respond(w, r, nil, http.StatusNotImplemented)
	}
}

func (s *userService) handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.Respond(w, r, "ping", http.StatusOK)
	}
}
