package rmx

import (
	"context"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rog-golang-buddies/req"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	Port   string
	Router http.ServeMux
}

func (s *Server) initRoutes() {
	s.Router.HandleFunc("/", req.CheckMethod([]string{req.GET}, func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello world"))
		if err != nil {
			log.Printf("write failed: %v", err.Error())
		}
	}))
}

func (s *Server) ServeHTTP() {
	s.initRoutes()

	// don't use "*" as AllowedOrigins, new origins should be added explicitly
	c := cors.Options{
		AllowedOrigins: []string{"http://localhost:3000"},
	}
	handler := cors.New(c).Handler(&s.Router)

	serverCtx, serverStop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer serverStop()

	server := http.Server{
		Addr:         s.Port,
		Handler:      handler,
		ErrorLog:     log.Default(),     // set the logger for the server
		ReadTimeout:  10 * time.Second,  // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
		BaseContext: func(_ net.Listener) context.Context {
			return serverCtx
		},
	}

	g, gCtx := errgroup.WithContext(serverCtx)
	g.Go(func() error {
		// Run the server
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Printf("exit reason: %s \n", err)
	}
}
