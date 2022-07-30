package rmx

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rog-golang-buddies/req"
	"github.com/rs/cors"
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

	server := http.Server{
		Addr:         s.Port,
		Handler:      handler,
		ErrorLog:     log.Default(),     // set the logger for the server
		ReadTimeout:  10 * time.Second,  // max time to read request from the client
		WriteTimeout: 10 * time.Second,  // max time to write response to the client
		IdleTimeout:  120 * time.Second, // max time for connections using TCP Keep-Alive
	}

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal("server graceful shutdown timed out. forcing exit.")
			}
		}()

		go func() {
			// Run the server
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Fatalf("server closing: %v", err.Error())
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatalf("server shutdown failed: %v", err.Error())
		}
		serverStopCtx()

		defer cancel()
	}()

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
