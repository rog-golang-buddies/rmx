package service_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	srv "github.com/rog-golang-buddies/rmx/internal/auth/http"
	token "github.com/rog-golang-buddies/rmx/internal/auth/redis/v1"
	"github.com/rog-golang-buddies/rmx/test"

	"github.com/hyphengolang/prelude/testing/is"
)

const applicationJson = "application/json"

func TestAuthService(t *testing.T) {
	is := is.New(t)

	var authServer *httptest.Server
	{
		tc := token.NewClient()
		h := srv.NewService(test.NewCredentialsRepo(), srv.WithTokenClient(tc))
		authServer = httptest.NewServer(h)

		t.Cleanup(func() { authServer.Close() })
	}

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
