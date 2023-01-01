package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	service "github.com/rog-golang-buddies/rmx/services/user/http"
	repo "github.com/rog-golang-buddies/rmx/services/user/postgres"

	"github.com/hyphengolang/prelude/testing/is"
)

const (
	applicationJson = "application/json"

	dbName         = "user"
	pgUser         = "postgres"
	pgPass         = "postgres"
	dbPort         = "5432"
	occurrence     = 2
	startUpTimeout = 5 * time.Second
)

func TestService(t *testing.T) {
	is := is.New(t)

	ctx := context.Background()

	container, conn,
		err := repo.NewUserDatabase(ctx, dbPort, pgUser, pgPass, dbName, occurrence, startUpTimeout)
	is.NoErr(err) // create new instance of the postgres container

	db := repo.New(conn)
	defer conn.Close()

	handler := service.NewService(db)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	t.Cleanup(func() {
		err := container.Terminate(ctx)
		is.NoErr(err) // terminate the container
	})

	t.Run("ping service", func(t *testing.T) {
		res, err := srv.Client().Get(srv.URL + "/api/v1/user/ping")
		is.NoErr(err)                           // ping server
		is.Equal(res.StatusCode, http.StatusOK) // return ok status
	})

	t.Run("register user", func(t *testing.T) {
		payload := `
		{
			"email":"fizz@gmail.com",
			"username":"fizz_user",
			"password":"fizz_$PW_10"
		}`

		res, _ := srv.Client().Post(srv.URL+"/api/v1/user/register", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusCreated) // new user created
	})
}
