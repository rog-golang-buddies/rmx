package sql

import (
	"context"
)

type RWRepo[Typ any] interface {
	// Context() context.Context
	Reader[Typ]
	Writer[Typ]
}

type Closer interface {
	Close() error
}

type Reader[Typ any] interface {
	Read(ctx context.Context, key any) (*Typ, error)
	ReadAll(ctx context.Context) ([]Typ, error)
}

type Writer[Typ any] interface {
	Write(ctx context.Context, v *Typ) error
	Remove(ctx context.Context, key any) error
}
