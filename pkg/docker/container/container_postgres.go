package container

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rog-golang-buddies/rmx/pkg/docker/options"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer represents the postgres container type used in the module
type PostgresContainer struct {
	testcontainers.Container
}

// setupPostgres creates an instance of the postgres container type
func NewPostgres(ctx context.Context, opts ...options.ContainerOption) (*PostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:11-alpine",
		Env:          map[string]string{},
		ExposedPorts: []string{},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	return &PostgresContainer{Container: container}, nil
}

func (c *PostgresContainer) ParseConnStr(ctx context.Context, port nat.Port, user, password, dbName string) (connStr string, err error) {
	var portC nat.Port
	if portC, err = c.MappedPort(ctx, port); err != nil {
		return
	}

	var hostC string
	if hostC, err = c.Host(ctx); err != nil {
		return "", err

	}

	connStr = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", hostC, portC.Port(), user, password, dbName)
	return
}

// NewPostgresConnection creates a new postgres container and returns a connection pool
func NewDefaultPostgres(ctx context.Context, port string, user, password, dbName string, occurrence int, startUpTimeout time.Duration) (*PostgresContainer, *pgxpool.Pool, error) {
	p, err := nat.NewPort("tcp", port)
	if err != nil {
		return nil, nil, err
	}

	c, err := NewPostgres(ctx,
		options.WithPort(p.Port()),
		options.WithInitialDatabase(user, password, dbName),
		options.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(occurrence).WithStartupTimeout(startUpTimeout)),
	)
	if err != nil {
		return nil, nil, err
	}

	connStr, err := c.ParseConnStr(ctx, p, user, password, dbName)
	if err != nil {
		return nil, nil, err
	}

	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	return c, conn, err
}

func NewPostgresContainer(ctx context.Context, internal nat.Port, opts ...options.ContainerOption) (*PostgresContainer, *pgxpool.Pool, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:11-alpine",
		Env:          map[string]string{},
		ExposedPorts: []string{internal.Port()},
		Cmd:          []string{"postgres", "-c", "fsync=off"},
	}

	for _, opt := range opts {
		opt(&req)
	}

	if len(req.ExposedPorts) != 1 {
		return nil, nil, fmt.Errorf("exposed ports must be exactly 1")
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{ContainerRequest: req, Started: true})
	if err != nil {
		return nil, nil, err
	}

	postgresContainer := &PostgresContainer{Container: container}

	externalPort, err := postgresContainer.MappedPort(ctx, internal)
	if err != nil {
		return nil, nil, err
	}

	externalHost, err := postgresContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		externalHost, externalPort.Port(), req.Env["POSTGRES_USER"], req.Env["POSTGRES_PASSWORD"], req.Env["POSTGRES_DB"])

	conn, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, nil, err
	}

	return postgresContainer, conn, nil
}
