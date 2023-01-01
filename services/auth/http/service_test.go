package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/rog-golang-buddies/rmx/services/auth"
	service "github.com/rog-golang-buddies/rmx/services/auth/http"
	repo "github.com/rog-golang-buddies/rmx/services/auth/postgres"
	token "github.com/rog-golang-buddies/rmx/services/auth/redis/v1"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/prelude/types/email"
)

const applicationJson = "application/json"

func TestAuthService(t *testing.T) {
	is := is.New(t)

	var authServer *httptest.Server
	t.Run("init service", func(t *testing.T) {
		tc := token.NewClient()
		h := service.NewService(mockCredentialsRepo(), service.WithTokenClient(tc))
		authServer = httptest.NewServer(h)
	})

	t.Cleanup(func() {
		authServer.Close()
	})

	t.Run("register new user's credentials", func(t *testing.T) {
		payload := `
		{
			"email":"fizz@gmail.com",
			"password":"fizz_$PW_10"
		}`

		res, _ := authServer.Client().Post(authServer.URL+"/api/v1/auth/credentials", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusNoContent) // status created
	})

	var accessToken string
	t.Run("user login with credentials", func(t *testing.T) {
		payload := `
		{
			"email":"fizz@gmail.com",
			"password":"fizz_$PW_10"
		}`

		res, _ := authServer.Client().Post(authServer.URL+"/api/v1/auth/login", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusOK) // login user with credentials

		type body struct {
			AccessToken string `json:"accessToken"`
		}

		var b body
		err := json.NewDecoder(res.Body).Decode(&b)
		res.Body.Close()
		is.NoErr(err) // parsing json

		// use alias to clean this up?
		accessToken = b.AccessToken

		// 		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/account/me", nil)
		// 		res, _ = srv.Client().Do(req)
		// 		is.Equal(res.StatusCode, http.StatusOK) // authorized endpoint
	})

	t.Run("verify user is logged in", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, authServer.URL+"/api/v1/auth/credentials", nil)
		req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, accessToken))
		res, _ := authServer.Client().Do(req)
		is.Equal(res.StatusCode, http.StatusNoContent) // verify user is logged in
	})

	t.Run("user logout", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, authServer.URL+"/api/v1/auth/login", nil)
		res, _ := authServer.Client().Do(req)
		is.Equal(res.StatusCode, http.StatusNoContent) // delete cookie
	})
}

// var s http.Handler

// func init() {
// 	ctx, mux := context.Background(), chi.NewMux()

// 	s = NewService(ctx, mux, repotest.NewUserRepo(), auth.DefaultTokenClient)
// }

// func TestService(t *testing.T) {
// 	t.Parallel()
// 	is := is.New(t)

// 	srv := httptest.NewServer(s)
// 	t.Cleanup(func() { srv.Close() })

// 	t.Run("sign-in, access auth endpoint then sign-out", func(t *testing.T) {
// 		payload := `
// 		{
// 			"email":"fizz@gmail.com",
// 			"password":"fizz_$PW_10"
// 		}`

// 		res, _ := srv.Client().
// 			Post(srv.URL+"/api/v1/auth/sign-in", applicationJson, strings.NewReader(payload))
// 		is.Equal(res.StatusCode, http.StatusOK)

// 		type body struct {
// 			IDToken     string `json:"idToken"`
// 			AccessToken string `json:"accessToken"`
// 		}

// 		var b body
// 		err := json.NewDecoder(res.Body).Decode(&b)
// 		res.Body.Close()
// 		is.NoErr(err) // parsing json

// 		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/account/me", nil)
// 		req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, b.AccessToken))
// 		res, _ = srv.Client().Do(req)
// 		is.Equal(res.StatusCode, http.StatusOK) // authorized endpoint

// 		req, _ = http.NewRequest(http.MethodDelete, srv.URL+"/api/v1/auth/sign-out", nil)
// 		req.Header.Set(`Authorization`, fmt.Sprintf(`Bearer %s`, b.AccessToken))
// 		res, _ = srv.Client().Do(req)
// 		is.Equal(res.StatusCode, http.StatusOK) // delete cookie
// 	})

// 	t.Run("refresh token", func(t *testing.T) {
// 		payload := `
// 		{
// 			"email":"fizz@gmail.com",
// 			"password":"fizz_$PW_10"
// 		}`

// 		res, _ := srv.Client().
// 			Post(srv.URL+"/api/v1/auth/sign-in", applicationJson, strings.NewReader(payload))
// 		is.Equal(res.StatusCode, http.StatusOK) // add refresh token

// 		// get the refresh token from the response's `Set-Cookie` header
// 		c := &http.Cookie{}
// 		for _, k := range res.Cookies() {
// 			t.Log(k.Value)
// 			if k.Name == cookieName {
// 				c = k
// 			}
// 		}

// 		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/v1/auth/refresh", nil)
// 		req.AddCookie(c)

// 		res, _ = srv.Client().Do(req)
// 		is.Equal(res.StatusCode, http.StatusOK) // refresh token
// 	})
// }

type credentialsRepo struct {
	mc sync.Map
}

// Read implements repo.CredentialsRepo
func (c *credentialsRepo) Read(ctx context.Context, key any) (*auth.Credentials, error) {
	if email, ok := key.(email.Email); !ok {
		return nil, fmt.Errorf("invalid key type")
	} else {
		if creds, ok := c.mc.Load(email); ok {
			return creds.(*auth.Credentials), nil
		} else {
			return nil, fmt.Errorf("credentials not found")
		}
	}
}

// ReadAll implements repo.CredentialsRepo
func (*credentialsRepo) ReadAll(ctx context.Context) ([]auth.Credentials, error) {
	panic("unimplemented")
}

// Remove implements repo.CredentialsRepo
func (*credentialsRepo) Remove(ctx context.Context, key any) error {
	panic("unimplemented")
}

// Write implements repo.CredentialsRepo
func (c *credentialsRepo) Write(ctx context.Context, creds *auth.Credentials) error {
	if _, ok := c.mc.Load(creds.Email); ok {
		return fmt.Errorf("user already exists")
	} else {
		c.mc.Store(creds.Email, creds)
	}

	return nil
}

func mockCredentialsRepo() repo.CredentialsRepo {
	c := &credentialsRepo{
		mc: sync.Map{},
	}

	return c
}
