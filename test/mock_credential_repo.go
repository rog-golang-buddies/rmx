package test

import (
	"context"
	"fmt"
	"sync"

	"github.com/hyphengolang/prelude/types/email"
	"github.com/rog-golang-buddies/rmx/internal/auth"
	repo "github.com/rog-golang-buddies/rmx/internal/auth/postgres"
)

var (
	ErrNotImplemented = fmt.Errorf("not implemented")
)

type credentialsRepo struct {
	mc sync.Map
}

// Read implements repo.CredentialsRepo
func (c *credentialsRepo) Read(ctx context.Context, key any) (*auth.Credentials, error) {
	if email, ok := key.(email.Email); !ok {
		return nil, fmt.Errorf("invalid key type")
	} else {
		if creds, ok := c.mc.Load(email); ok {
			return creds.(*auth.Credentials), nil
		} else {
			return nil, fmt.Errorf("credentials not found")
		}
	}
}

// ReadAll implements repo.CredentialsRepo
func (*credentialsRepo) ReadAll(ctx context.Context) ([]auth.Credentials, error) {
	return nil, ErrNotImplemented
}

// Remove implements repo.CredentialsRepo
func (*credentialsRepo) Remove(ctx context.Context, key any) error {
	return ErrNotImplemented
}

// Write implements repo.CredentialsRepo
func (c *credentialsRepo) Write(ctx context.Context, creds *auth.Credentials) error {
	if _, ok := c.mc.Load(creds.Email); ok {
		return fmt.Errorf("user already exists")
	} else {
		c.mc.Store(creds.Email, creds)
	}

	return nil
}

func NewCredentialsRepo() repo.CredentialsRepo {
	c := &credentialsRepo{
		mc: sync.Map{},
	}

	return c
}
