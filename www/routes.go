package www

import "net/http"

// This can be used as a ping-pong test to check if the server is up and running.
//  GET /ping
func (s Service) handlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, "pong", http.StatusOK)
	}
}

func (s Service) routes() {
	s.m.Handle("/assets/*", s.fileServer("/assets/", "assets"))

	s.m.Get("/ping", s.handlePing())

	s.m.Get("/ws/jam/{id}", s.handleTransmitMIDI())
	s.m.Get("/ws/signal/{id}", s.handleP2PSignal())
}

// Transmitting MIDI and other Jam Session-specific messages between musicians
//  GET /ws/jam/{id}
func (s Service) handleTransmitMIDI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.notImplemented(w, r)
	}
}

// Initialize WebRTC connections between peers in a specific Jam Session
//  GET /ws/signal/{id}
func (s Service) handleP2PSignal() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.notImplemented(w, r)
	}
}
