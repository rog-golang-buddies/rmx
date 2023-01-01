package repo_test

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/rog-golang-buddies/rmx/docker/container"
	"github.com/rog-golang-buddies/rmx/docker/options"
	"github.com/rog-golang-buddies/rmx/services/auth"
	repo "github.com/rog-golang-buddies/rmx/services/auth/postgres"

	"github.com/docker/go-connections/nat"
	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	//go:embed init/init.sql
	migration string

	dockerContainer *container.PostgresContainer
	postgresPool    *pgxpool.Pool
)

func TestRepo(t *testing.T) {
	is := is.New(t)

	t.Run("migration setup", func(t *testing.T) {
		ctx := context.Background()

		port, err := nat.NewPort("tcp", "5432")
		is.NoErr(err) // create new instance of the port

		dockerContainer, postgresPool, err = container.NewPostgresContainer(
			ctx,
			port,
			options.WithInitialDatabase("auth", "auth", "auth"),
			options.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
		)
		is.NoErr(err) // create new instance of the postgres container

		_, err = postgresPool.Exec(ctx, migration)
		is.NoErr(err) // run the migration
	})

	t.Cleanup(func() {
		ctx := context.Background()

		err := dockerContainer.Terminate(ctx)
		is.NoErr(err) // terminate the container
	})

	// new auth repo
	repo := repo.New(postgresPool)

	t.Run("insert new credentials", func(t *testing.T) {
		ctx := context.Background()

		cred := auth.Credentials{
			Email:    email.MustParse("fizz@gmail.com"),
			Password: password.MustParse("pA$4w()rD").MustHash(),
		}

		err := repo.Write(ctx, &cred)
		is.NoErr(err) // insert new credentials
	})

	t.Run("prevent duplicates", func(t *testing.T) {
		ctx := context.Background()

		cred := auth.Credentials{
			Email:    email.MustParse("fizz@gmail.com"),
			Password: password.MustParse("pA4$w0rD").MustHash(),
		}

		err := repo.Write(ctx, &cred)
		is.True(err != nil) // prevent duplicates
	})

	t.Run("fetch credentials based on email address", func(t *testing.T) {
		ctx := context.Background()

		cred, err := repo.Read(ctx, email.MustParse("fizz@gmail.com"))
		is.NoErr(err) // fetch credentials based on email address

		is.Equal(cred.Email.String(), "fizz@gmail.com") // check email address
	})

	t.Run("fetch credentials based on email address (not found)", func(t *testing.T) {
		ctx := context.Background()

		_, err := repo.Read(ctx, email.MustParse("buzz@gmail.com"))
		is.True(err != nil) // fetch credentials based on email address (not found)
	})
}
