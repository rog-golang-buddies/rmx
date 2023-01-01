package auth

import (
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
)

type Credentials struct {
	Email    email.Email           `json:"email"`
	Password password.PasswordHash `json:"-"`
	// TODO - likely more info that is required for auth services
}
