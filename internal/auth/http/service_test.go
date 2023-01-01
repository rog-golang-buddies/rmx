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

	intern "github.com/rog-golang-buddies/rmx/internal/auth"
	srv "github.com/rog-golang-buddies/rmx/internal/auth/http"
	repo "github.com/rog-golang-buddies/rmx/internal/auth/postgres"
	token "github.com/rog-golang-buddies/rmx/internal/auth/redis/v1"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/prelude/types/email"
)

const applicationJson = "application/json"

func TestAuthService(t *testing.T) {
	is := is.New(t)

	var authServer *httptest.Server
	t.Run("init service", func(t *testing.T) {
		tc := token.NewClient()
		h := srv.NewService(mockCredentialsRepo(), srv.WithTokenClient(tc))
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

type credentialsRepo struct {
	mc sync.Map
}

// Read implements repo.CredentialsRepo
func (c *credentialsRepo) Read(ctx context.Context, key any) (*intern.Credentials, error) {
	if email, ok := key.(email.Email); !ok {
		return nil, fmt.Errorf("invalid key type")
	} else {
		if creds, ok := c.mc.Load(email); ok {
			return creds.(*intern.Credentials), nil
		} else {
			return nil, fmt.Errorf("credentials not found")
		}
	}
}

// ReadAll implements repo.CredentialsRepo
func (*credentialsRepo) ReadAll(ctx context.Context) ([]intern.Credentials, error) {
	panic("unimplemented")
}

// Remove implements repo.CredentialsRepo
func (*credentialsRepo) Remove(ctx context.Context, key any) error {
	panic("unimplemented")
}

// Write implements repo.CredentialsRepo
func (c *credentialsRepo) Write(ctx context.Context, creds *intern.Credentials) error {
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
