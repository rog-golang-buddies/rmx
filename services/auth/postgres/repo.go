package repo

import (
	"context"
	"fmt"

	"github.com/rog-golang-buddies/rmx/common/sql"
	"github.com/rog-golang-buddies/rmx/services/auth"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidKey     = fmt.Errorf("invalid key type")
	ErrNotImplemented = fmt.Errorf("not implemented")
)

type Credentials struct {
	ID       int
	Email    email.Email
	Password password.PasswordHash
}

type repo struct {
	rh sql.PSQLHandler[Credentials]
}

// Read implements sql.RWRepo
func (r *repo) Read(ctx context.Context, key any) (*auth.Credentials, error) {
	// key is likely only going to be of typ e`email.Email` going forth
	if email, ok := key.(email.Email); !ok {
		return nil, ErrInvalidKey
	} else {
		var c auth.Credentials
		err := r.rh.QueryRow(ctx, qrySelectViaEmail, func(r pgx.Row) error { return r.Scan(&c.Email, &c.Password) }, email)
		return &c, err
	}
}

// ReadAll implements sql.RWRepo
func (*repo) ReadAll(ctx context.Context) ([]auth.Credentials, error) {
	return nil, ErrNotImplemented
}

// Remove implements sql.RWRepo
func (r *repo) Remove(ctx context.Context, key any) error {
	// key is likely only going to be of typ e`email.Email` going forth
	if email, ok := key.(email.Email); !ok {
		return ErrInvalidKey
	} else {
		// NOTE - somewhere in the pipeline a password validation may be required
		err := r.rh.Exec(ctx, qrySelectViaEmail, email)
		return err
	}
}

// Write implements sql.RWRepo
func (r *repo) Write(ctx context.Context, c *auth.Credentials) error {
	// NOTE - it is ok to use auth.Credentials type here
	// as there isn't much mapping going on between the types
	return r.rh.Exec(ctx, qryInsert, c.Email, c.Password)
}

type CredentialsRepo sql.RWRepo[auth.Credentials]

func New(conn *pgxpool.Pool, opts ...Option) CredentialsRepo {
	r := &repo{
		rh: sql.NewPSQLHandler[Credentials](conn),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type Option func(*repo)

const (
	qryInsert = `INSERT INTO "credentials" (email, password) VALUES ($1, $2);`

	qrySelectViaEmail = `SELECT email, password FROM "credentials" WHERE email = $1;`

	qryDeleteViaEmail = `DELETE FROM "credentials" WHERE email = $1;`
)
