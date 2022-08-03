package www

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gotest.tools/assert"
)

func TestRoutes(t *testing.T) {
	type testcase struct {
		Name        string
		PathName    string
		Method      string
		ContentType string
		Content     string
		Status      int
	}

	tt := []testcase{
		// {Name: "serving assets", PathName: "/assets/data.json", Method: "GET", Status: http.StatusOK}, // httptest can't test this
		{Name: "ping pong test", PathName: "/ping", Method: "GET", Status: http.StatusOK},
		{Name: "transmitting midi", PathName: "/ws/jam/0", Method: "GET", Status: http.StatusNotImplemented},
		{Name: "creating p2p connection", PathName: "/ws/signal/0", Method: "GET", Status: http.StatusNotImplemented},
	}

	srv := httptest.NewServer(NewService())

	for _, tc := range tt {
		if tc.ContentType == "" {
			tc.ContentType = "application/json"
		}
		if tc.Status == 0 {
			tc.Status = http.StatusOK
		}
		t.Run(tc.Name, func(t *testing.T) {
			req, err := http.NewRequest(tc.Method, srv.URL+tc.PathName, strings.NewReader(tc.Content))
			assert.Equal(t, err, nil)

			req.Header.Add("Content-Type", tc.ContentType)

			res, err := srv.Client().Do(req)
			assert.Equal(t, err, nil)

			assert.Equal(t, res.StatusCode, tc.Status)
		})
	}
}
