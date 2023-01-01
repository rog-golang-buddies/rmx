package service

import (
	"fmt"
	"net/http"
	"time"

	intern "github.com/rog-golang-buddies/rmx/internal/auth"
	repo "github.com/rog-golang-buddies/rmx/internal/auth/postgres"
	token "github.com/rog-golang-buddies/rmx/internal/auth/redis/v1"
	srv "github.com/rog-golang-buddies/rmx/internal/http"
	"github.com/rog-golang-buddies/rmx/pkg/auth"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type authService struct {
	mux srv.Service

	tc *token.Client

	r repo.CredentialsRepo
}

func (s *authService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func NewService(credRepo repo.CredentialsRepo, opts ...AuthOption) http.Handler {
	// TODO - use builder pattern instead
	s := &authService{mux: srv.New(), r: credRepo}
	for _, opt := range opts {
		opt(s)
	}

	if s.r == nil {
		panic("credentials repo is nil")
	}

	if s.tc == nil {
		panic("token client is nil")
	}

	s.routes()
	return s
}

func (s *authService) routes() {
	// TODO - replace with tokenClient
	public, private, _ := s.tc.GenerateKeys()

	s.mux.Post("/api/v1/auth/credentials", s.handleCreateCredentials())
	// NOTE - this endpoint just authenticates the user
	s.mux.Get("/api/v1/auth/credentials", s.handleVerify(public))

	s.mux.Post("/api/v1/auth/login", s.handleLogin(private))
	s.mux.Delete("/api/v1/auth/login", s.handleLogout())

}

func (s *authService) handleVerify(public jwk.Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.authenticate(w, r, public); err != nil {
			s.mux.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		s.mux.Respond(w, r, nil, http.StatusNoContent)
	}
}

func (s *authService) handleCreateCredentials() http.HandlerFunc {
	type payload struct {
		Email    email.Email       `json:"email"`
		Password password.Password `json:"password"`
	}

	newCreds := func(w http.ResponseWriter, r *http.Request, creds *intern.Credentials) error {
		var dto payload
		if err := s.mux.Decode(w, r, &dto); err != nil {
			return err
		}

		// TODO - hash password
		h, err := dto.Password.Hash()
		if err != nil {
			return err
		}

		*creds = intern.Credentials{
			Email:    dto.Email,
			Password: h,
		}

		return nil
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var creds intern.Credentials
		if err := newCreds(w, r, &creds); err != nil {
			s.mux.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		if err := s.r.Write(r.Context(), &creds); err != nil {
			s.mux.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		s.mux.Respond(w, r, nil, http.StatusNoContent)
	}
}

func (s *authService) handleLogin(private jwk.Key) http.HandlerFunc {
	type payload struct {
		Email    email.Email       `json:"email"`
		Password password.Password `json:"password"`
	}

	type response struct {
		IDToken     string `omitempty,json:"idToken"`
		AccessToken string `omitempty,json:"accessToken"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		var dto payload
		if err := s.mux.Decode(w, r, &dto); err != nil {
			s.mux.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		cred, err := s.r.Read(r.Context(), dto.Email)
		if err != nil {
			s.mux.Respond(w, r, err, http.StatusNotFound)
			return
		}

		if err := cred.Password.Compare(dto.Password.String()); err != nil {
			s.mux.Respond(w, r, err, http.StatusUnauthorized)
			return
		}

		opt := auth.TokenOption{
			// NOTE - this could be a variable of the token client
			Issuer: issuer,
			// NOTE - need to replace u.UUID with a client based ID
			// this will mean different cookies for multi-device usage
			Subject: suid.NewUUID().ShortUUID().String(), // new client ID for tracking user connections
			// Audience: []string{},
			Claims: map[string]any{"email": cred.Email},
		}

		its, ats, rts, err := s.signedTokens(private, &opt)
		if err != nil {
			s.mux.Respond(w, r, err, http.StatusInternalServerError)
			return
		}

		c := s.newCookie(w, r, string(rts), auth.RefreshTokenExpiry)

		tk := &response{
			IDToken:     string(its),
			AccessToken: string(ats),
		}

		s.mux.SetCookie(w, c)
		s.mux.Respond(w, r, tk, http.StatusOK)
	}
}

func (s *authService) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := s.newCookie(w, r, "", -1)

		s.mux.SetCookie(w, c)
		s.mux.Respond(w, r, nil, http.StatusNoContent)
	}
}

// // FIXME this endpoint is broken due to the redis client
// // We need to try fix this ASAP
// func (s *Service) handleRefresh(public, private jwk.Key) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// // NOTE temp switch away from auth middleware
// jtk, err := auth.ParseCookie(r, public, cookieName)
// if err != nil {
// 	s.Respond(w, r, err, http.StatusUnauthorized)
// 	return
// }

// claim, ok := jtk.PrivateClaims()["email"].(string)
// if !ok {
// 	s.RespondText(w, r, http.StatusInternalServerError)
// 	return
// }

// u, err := s.r.Select(r.Context(), email.Email(claim))
// if err != nil {
// 	s.Respond(w, r, err, http.StatusForbidden)
// 	return
// }

// // FIXME commented out as not complete
// // // already checked in auth but I am too tired
// // // to come up with a cleaner solution
// // k, _ := r.Cookie(cookieName)

// // err := s.tc.ValidateRefreshToken(r.Context(), k.Value)
// // if err != nil {
// // 	s.Respond(w, r, err, http.StatusInternalServerError)
// // 	return
// // }

// // // token validated, now it should be set inside blacklist
// // // this prevents token reuse
// // err = s.tc.BlackListRefreshToken(r.Context(), k.Value)
// // if err != nil {
// // 	s.Respond(w, r, err, http.StatusInternalServerError)
// // }

// // cid := j.Subject()
// // _, ats, rts, err := s.signedTokens(private, claim.String(), suid.SUID(cid))
// // if err != nil {
// // 	s.Respond(w, r, err, http.StatusInternalServerError)
// // 	return
// // }

// u.ID, _ = suid.ParseString(jtk.Subject())

// _, ats, rts, err := s.signedTokens(private, u)
// if err != nil {
// 	s.Respond(w, r, err, http.StatusInternalServerError)
// 	return
// }

// c := s.newCookie(w, r, string(rts), auth.RefreshTokenExpiry)

// tk := &Token{
// 	AccessToken: string(ats),
// }

// s.SetCookie(w, c)
// s.Respond(w, r, tk, http.StatusOK)
// 	}
// }

func (s *authService) newCookie(w http.ResponseWriter, r *http.Request, value string, maxAge time.Duration) *http.Cookie {
	c := &http.Cookie{
		Path:     "/",
		Name:     cookieName,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(maxAge),
		Value:    string(value),
	}
	return c
}

func (s *authService) authenticate(w http.ResponseWriter, r *http.Request, public jwk.Key) error {
	tk, err := auth.ParseRequest(r, public)
	if err != nil {
		s.mux.Logf("token: %v", tk)
		return err
	}

	claim, ok := tk.PrivateClaims()["email"].(string)
	if err := fmt.Errorf("email claim does not exist"); !ok {
		return err
	}

	// NOTE - an assumption can be made that this email will parse.
	// Unsure if rest of the team would agree
	if _, err = s.r.Read(r.Context(), email.MustParse(claim)); err != nil {
		return err
	}

	return nil
}

// // TODO there is two cid's being used here, need clarification
func (s *authService) signedTokens(key jwk.Key, opt *auth.TokenOption) (its, ats, rts []byte, err error) {
	opt.Expiration = s.tc.TokenExp["id"]
	if its, err = s.tc.GenerateToken(key, opt); err != nil {
		return
	}

	opt.Expiration = s.tc.TokenExp["access"]
	if ats, err = s.tc.GenerateToken(key, opt); err != nil {
		return
	}

	opt.Expiration = s.tc.TokenExp["refresh"]
	if rts, err = s.tc.GenerateToken(key, opt); err != nil {
		return
	}

	return
}

const (
	// FIXME - move this to TokenClient
	issuer     = "github.com/rog-golang-buddies/rmx"
	cookieName = "RMX_REFRESH_TOKEN"
)

type AuthOption srv.Option[*authService]

// WithCredentialsRepo sets the credentials repo
func WithCredentialsRepo(r repo.CredentialsRepo) AuthOption {
	return func(s *authService) {
		s.r = r
	}
}

// WithTokenClient sets the token client
func WithTokenClient(tc *token.Client) AuthOption {
	return func(s *authService) {
		s.tc = tc
	}
}
