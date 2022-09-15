package www

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"

	ws "github.com/rog-golang-buddies/rapidmidiex/www/ws"
)

func (s Service) sessionPool(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, err := s.parseUUID(w, r, "id")
		if err != nil {
			s.respond(w, r, err, http.StatusBadRequest)
			return
		}

		p, err := s.c.Get(uid)
		if err != nil {
			s.l.Println(err)
			return
		}

		s.l.Println("This session matches the ID", p.ID)

		r = r.WithContext(context.WithValue(r.Context(), roomKey, p))
		f(w, r)
	}
}

func (s Service) upgradeHTTP(f http.HandlerFunc) http.HandlerFunc {
	u := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		p := r.Context().Value(roomKey).(*ws.Pool)

		if p.Size() == p.MaxConn {
			s.respond(w, r, ws.ErrMaxConn, http.StatusUnauthorized)
			return
		}

		c, err := p.NewConn(w, r, u)
		if err != nil {
			s.respond(w, r, err, http.StatusInternalServerError)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), upgradeKey, c))
		f(w, r)
	}
}
