package ws

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Conn struct {
	ID uuid.UUID

	rwc *websocket.Conn
	p   *Pool
}

func (c Conn) Pool() *Pool { return c.p }

func (c Conn) Close() error {
	c.p.Delete(c.ID)

	return c.rwc.Close()
}

func (c Conn) ReadJSON(v any) error { return c.rwc.ReadJSON(v) }

func (c Conn) WriteJSON(v any) error { return c.rwc.WriteJSON(v) }

func (c Conn) SendMessage(v any) error { c.p.msgs <- v; return nil }
