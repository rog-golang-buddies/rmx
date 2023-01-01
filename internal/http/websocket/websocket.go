package websocket

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/gobwas/ws/wsutil"
	"github.com/hyphengolang/prelude/types/suid"
)

type Err[CI any] struct {
	conn *Conn[CI]
	msg  error
}

func (e *Err[CI]) Error() string {
	return e.msg.Error()
}

// A Web-Socket Connection
type Conn[CI any] struct {
	sid  suid.UUID
	rwc  io.ReadWriteCloser
	lock sync.RWMutex

	Info *CI
}

// Writes raw bytes to the Connection
func (c *Conn[CI]) write(b []byte) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return wsutil.WriteServerBinary(c.rwc, b)
}

type WSMsgTyp int

const (
	Text WSMsgTyp = iota + 1
	JSON
	Leave
)

// type for parsing bytes into messages
type message struct {
	typ  WSMsgTyp
	data []byte
}

// Parses the bytes into the message type
func (m *message) parse(b []byte) {
	// the first byte represents the data type (Text, JSON, Leave)
	m.typ = WSMsgTyp(b[0])
	// and others represent the data itself
	m.data = b[1:]
}

func (m *message) marshall() []byte {
	return append([]byte{byte(m.typ)}, m.data...)
}

// Converts the given bytes to string
func (m *message) readText() (string, error) {
	return string(m.data), nil
}

// Converts the given bytes to JSON
func (m *message) readJSON() (any, error) {
	var v any
	if err := json.Unmarshal(m.data, v); err != nil {
		return nil, err
	}

	return v, nil
}

type Reader interface {
	ReadText() (string, error)
	ReadJSON() (interface{}, error)
}

type Writer interface {
	WriteText(s string)
	WriteJSON(i any) error
}
