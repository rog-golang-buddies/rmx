package service_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	srv "github.com/rog-golang-buddies/rmx/internal/user/http"
	"github.com/rog-golang-buddies/rmx/test"

	h "github.com/hyphengolang/prelude/http"
	"github.com/hyphengolang/prelude/testing/is"
	"github.com/sergi/go-diff/diffmatchpatch"
)

const (
	applicationJson = "application/json"
)

// https://github.com/OperationSpark/service-signups/blob/main/zoom_test.go
//
// https://github.com/googlemaps/google-maps-services-go/blob/e6c76e578df330b3985c79df687a6c2953f450f3/directions_test.go#L41
func TestService(t *testing.T) {
	is := is.New(t)

	authServer := mockServer(http.StatusCreated, "OK")

	var userServer *httptest.Server
	{
		h := srv.NewService(
			srv.WithRepo(test.NewUserRepo()),
			srv.WithAuthServiceURL(authServer.URL),
		)

		userServer = httptest.NewServer(h)

		t.Cleanup(func() { userServer.Close() })
	}

	t.Run("ping service", func(t *testing.T) {
		res, err := userServer.Client().Get(userServer.URL + "/api/v1/user/ping")
		is.NoErr(err)                                  // ping server
		is.Equal(res.StatusCode, http.StatusNoContent) // return ok status
	})

	t.Run("register some users", func(t *testing.T) {
		payload := `
		{
			"email":"fizz@gmail.com",
			"username":"fizz_user",
			"password":"fizz_$PW_10"
		}`

		res, _ := userServer.Client().Post(userServer.URL+"/api/v1/user/register", applicationJson, strings.NewReader(payload))
		is.Equal(res.StatusCode, http.StatusCreated) // new user created

		// payload = `
		// {
		// 	"email":"buzz@gmail.com",
		// 	"username":"buzz_user",
		// 	"password":"buzz_$PW_10"
		// }`

		// res, _ = userServer.Client().Post(userServer.URL+"/api/v1/user/register", applicationJson, strings.NewReader(payload))
		// is.Equal(res.StatusCode, http.StatusCreated) // new user created

		// payload = `
		// {
		// 	"email":"bazz@gmail.com",
		// 	"username":"buzz_user",
		// 	"password":"bazz_$PW_10"
		// }`

		// res, _ = userServer.Client().Post(userServer.URL+"/api/v1/user/register", applicationJson, strings.NewReader(payload))
		// is.Equal(res.StatusCode, http.StatusInternalServerError) // new user created
	})
}

// create a mock server to handle auth service requests
type authServer struct {
	s          *httptest.Server
	successful int
	failed     []string
}

func mockServerForQuery(query string, code int, body string) *authServer {
	server := &authServer{}

	server.s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if query != "" && r.URL.RawQuery != query {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(query, r.URL.RawQuery, false)
			log.Printf("Query != Expected Query: %s", dmp.DiffPrettyText(diffs))
			server.failed = append(server.failed, r.URL.RawQuery)
			http.Error(w, "fail", 999)
			return
		}
		server.successful++

		h.Respond(w, r, body, code)
	}))

	return server
}

func mockServer(code int, body string) *httptest.Server {
	serv := mockServerForQuery("", code, body)
	return serv.s
}
