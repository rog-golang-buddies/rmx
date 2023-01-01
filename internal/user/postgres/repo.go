package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/rog-golang-buddies/rmx/internal/user"
	"github.com/rog-golang-buddies/rmx/pkg/docker/container"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rog-golang-buddies/rmx/internal/sql"

	_ "embed"
)

//go:embed init/init.sql
var migration string

// NOTE - for testing purposes only
func NewUserDatabase(ctx context.Context, port string, user string, password string, dbName string, occurrence int, startUpTimeout time.Duration) (*container.PostgresContainer, *pgxpool.Pool, error) {
	c, pg, err := container.NewDefaultPostgres(ctx, port, user, password, dbName, occurrence, startUpTimeout)
	if err != nil {
		return nil, nil, err
	}

	_, err = pg.Exec(ctx, migration)
	return c, pg, err
}

var (
	ErrInvalidKey = fmt.Errorf("invalid key type")
)

type User struct {
	// Primary key.
	ID suid.UUID
	// Unique. Stored as text.
	Username string
	// Unique. Stored as case-sensitive text.
	Email email.Email
	// Required. Stored as case-sensitive text.
	Password password.PasswordHash
	// Required. Defaults to current time.
	CreatedAt time.Time
	// TODO nullable, currently inactive
	// UpdatedAt *time.Time
	// TODO nullable, currently inactive
	// DeletedAt *time.Time
}

func (r *repo) ReadAll(ctx context.Context) ([]user.User, error) {
	us, err := r.rh.Query(ctx, qrySelectMany, func(r pgx.Rows, u *user.User) error {
		return r.Scan(&u.ID, &u.Email, &u.Username, &u.Password)
	})

	return us, err
}

func (r *repo) Read(ctx context.Context, key any) (*user.User, error) {
	var qry string
	switch key.(type) {
	case suid.UUID:
		qry = qrySelectByID
	case email.Email:
		qry = qrySelectByEmail
	case string:
		qry = qrySelectByUsername
	default:
		return nil, ErrInvalidKey
	}

	var u user.User
	return &u, r.rh.QueryRow(ctx, qry, func(r pgx.Row) error { return r.Scan(&u.ID, &u.Username, &u.Email, &u.Password) }, key)
}

func (r *repo) Write(ctx context.Context, u *user.User) error {
	args := pgx.NamedArgs{
		"id":       u.ID,
		"email":    u.Email,
		"username": u.Username,
		"password": u.Password,
	}

	if err := r.rh.Exec(ctx, qryInsert, args); err != nil {
		r.rh.Logf("error writing user: %v", err)
		return err
	}

	return nil
}

func (r *repo) Remove(ctx context.Context, key any) error {
	var qry string
	switch key.(type) {
	case suid.UUID:
		qry = qryDeleteByID
	case email.Email:
		qry = qryDeleteByEmail
	case string:
		qry = qryDeleteByUsername
	default:
		return ErrInvalidKey
	}
	return r.rh.Exec(ctx, qry, key)
}

type UserRepo sql.RWRepo[user.User]

type repo struct {
	rh sql.PSQLHandler[user.User]
}

func New(conn *pgxpool.Pool) UserRepo {
	r := &repo{
		rh: sql.NewPSQLHandler[user.User](conn),
	}
	return r
}

const (
	qryInsert = `insert into "user" (id, email, username, password) values (@id, @email, @username, @password)`

	qrySelectMany = `select id, email, username, password from "user" order by id`

	qrySelectByID       = `select id, email, username, password from "user" where id = $1`
	qrySelectByEmail    = `select id, email, username, password from "user" where email = $1`
	qrySelectByUsername = `select id, email, username, password from "user" where username = $1`

	qryDeleteByID       = `delete from "user" where id = $1`
	qryDeleteByEmail    = `delete from "user" where email = $1`
	qryDeleteByUsername = `delete from "user" where username = $1`
)
