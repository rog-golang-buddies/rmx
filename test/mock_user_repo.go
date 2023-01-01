package test

import (
	"context"
	"fmt"
	"sync"

	"github.com/rog-golang-buddies/rmx/internal/user"
	repo "github.com/rog-golang-buddies/rmx/internal/user/postgres"
)

type userRepo struct {
	byId       sync.Map
	byEmail    sync.Map
	byUsername sync.Map
}

// Read implements postgres.UserRepo
func (*userRepo) Read(ctx context.Context, key any) (*user.User, error) {
	return nil, ErrNotImplemented
}

// ReadAll implements postgres.UserRepo
func (*userRepo) ReadAll(ctx context.Context) ([]user.User, error) {
	return nil, ErrNotImplemented

}

// Remove implements postgres.UserRepo
func (*userRepo) Remove(ctx context.Context, key any) error {
	return ErrNotImplemented

}

// Write implements postgres.UserRepo
func (r *userRepo) Write(ctx context.Context, v *user.User) error {
	_, hasEmail := r.byEmail.Load(v.Email)
	_, hasId := r.byId.Load(v.ID)
	_, hasUsername := r.byUsername.Load(v.Username)

	if hasEmail || hasId || hasUsername {
		return fmt.Errorf("user already exists")
	} else {
		/* store with ID & email index. */
		{
			r.byEmail.Store(v.Email, v)
			r.byId.Store(v.ID, v)
			r.byUsername.Store(v.Username, v)
		}
	}

	return nil
}

func NewUserRepo() repo.UserRepo {
	c := &userRepo{
		byId:       sync.Map{},
		byEmail:    sync.Map{},
		byUsername: sync.Map{},
	}

	return c
}
