package user

import (
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
)

type User struct {
	ID       suid.UUID         `json:"id"`
	Username string            `json:"username"`
	Email    email.Email       `json:"email"`
	Password password.Password `json:"-"`
}
