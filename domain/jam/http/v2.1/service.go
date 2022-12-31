package v2

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gobwas/ws"
	"github.com/hyphengolang/prelude/types/suid"

	service "github.com/rog-golang-buddies/rmx/common/http"
	"github.com/rog-golang-buddies/rmx/common/http/websocket"
	"github.com/rog-golang-buddies/rmx/domain/jam"
)

const (
	defaultTimeout = time.Second * 10
)

type jamService struct {
	mux service.Service

	wsb *websocket.Broker[jam.Jam, jam.User]
}

func (s *jamService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func NewService(ctx context.Context, cap uint) http.Handler {
	broker := websocket.NewBroker[jam.Jam, jam.User](ctx, cap)

	s := &jamService{
		mux: service.New(),
		wsb: broker,
	}

	s.routes()
	return s
}

func (s *jamService) handleCreateJamRoom() http.HandlerFunc {
	// NOTE - should be an intimidate Jam type before it is converted to a domain type
	return func(w http.ResponseWriter, r *http.Request) {
		var j jam.Jam
		if err := s.mux.Decode(w, r, &j); err != nil {
			s.mux.Respond(w, r, err, http.StatusBadRequest)
			return
		}

		sub := s.newSubscriber(&j)
		s.mux.Created(w, r, sub.GetID().ShortUUID().String())
	}
}

func (s *jamService) handleGetRoomData() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.Respond(w, r, nil, http.StatusNotImplemented)
	}
}

func (s *jamService) handleListRooms() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.mux.Respond(w, r, nil, http.StatusNotImplemented)
	}
}

func (s *jamService) handleP2PComms() http.HandlerFunc {
	// FIXME - move to websocket package
	var ErrCapacity = fmt.Errorf("subscriber has reached max capacity")

	return func(w http.ResponseWriter, r *http.Request) {
		// decode uuid from URL
		sid, err := s.parseUUID(w, r)
		if err != nil {
			s.mux.Respond(w, r, sid, http.StatusBadRequest)
			return
		}

		sub, err := s.wsb.GetSubscriber(sid)
		if err != nil {
			s.mux.Respond(w, r, err, http.StatusNotFound)
			return
		}

		if sub.IsFull() {
			s.mux.Respond(w, r, ErrCapacity, http.StatusServiceUnavailable)
			return
		}

		rwc, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			// NOTE - this we discovered isn't needed
			s.mux.Respond(w, r, err, http.StatusUpgradeRequired)
			return
		}

		// NOTE - not sure what this actually does,
		// should be coming from database
		u := jam.NewUser("")

		conn := sub.NewConn(rwc, u)
		sub.Subscribe(conn)
	}
}

func (s *jamService) routes() {
	s.mux.Route("/api/v1/jam", func(r chi.Router) {
		r.Get("/", s.handleListRooms())
		r.Get("/{uuid}", s.handleGetRoomData())
		r.Post("/", s.handleCreateJamRoom())
	})

	s.mux.Route("/ws/jam", func(r chi.Router) {
		r.Get("/{uuid}", s.handleP2PComms())
	})

}

func (s *jamService) parseUUID(w http.ResponseWriter, r *http.Request) (suid.UUID, error) {
	return suid.ParseString(chi.URLParam(r, "uuid"))
}

func (s *jamService) newSubscriber(j *jam.Jam) *websocket.Subscriber[jam.Jam, jam.User] {
	sub := websocket.NewSubscriber[jam.Jam, jam.User](
		s.wsb.Context, j.Capacity, 512, defaultTimeout, defaultTimeout, j,
	)

	s.wsb.Subscribe(sub)
	return sub
}
