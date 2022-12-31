package options

import (
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type ContainerOption func(req *testcontainers.ContainerRequest)

func WithWaitStrategy(strategies ...wait.Strategy) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		req.WaitingFor = wait.ForAll(strategies...).WithDeadline(1 * time.Minute)
	}
}

func WithPort(port string) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		req.ExposedPorts = append(req.ExposedPorts, port)
	}
}

func WithInitialDatabase(user string, password string, dbName string) ContainerOption {
	return func(req *testcontainers.ContainerRequest) {
		req.Env["POSTGRES_USER"] = user
		req.Env["POSTGRES_PASSWORD"] = password
		req.Env["POSTGRES_DB"] = dbName
	}
}
