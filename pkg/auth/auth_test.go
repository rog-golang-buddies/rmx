package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/rog-golang-buddies/rmx/pkg/auth"

	"github.com/hyphengolang/prelude/testing/is"
	"github.com/hyphengolang/prelude/types/email"
	"github.com/hyphengolang/prelude/types/password"
	"github.com/hyphengolang/prelude/types/suid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type TestUser struct {
	ID       suid.UUID
	Username string
	Email    email.Email
	Password password.PasswordHash
}

func TestToken(t *testing.T) {
	is := is.New(t)

	t.Run(`generate a token and sign`, func(t *testing.T) {
		_, private := auth.ES256()

		u := TestUser{
			ID:       suid.NewUUID(),
			Username: "fizz_user",
			Email:    "fizz@mail.com",
			Password: password.Password("492045rf-vf").MustHash(),
		}

		o := auth.TokenOption{
			Issuer:     "github.com/rog-golang-buddies/rmx",
			Subject:    suid.NewUUID().String(),
			Expiration: time.Hour * 10,
			Claims:     map[string]any{"email": u.Email},
		}

		_, err := auth.Sign(private, &o)
		is.NoErr(err) // sign id token

		o.Subject = u.ID.String()
		o.Expiration = auth.AccessTokenExpiry

		_, err = auth.Sign(private, &o)
		is.NoErr(err) // access token

		o.Expiration = auth.RefreshTokenExpiry
		_, err = auth.Sign(private, &o)
		is.NoErr(err) // refresh token
	})
}

func TestMiddleware(t *testing.T) {
	is := is.New(t)

	t.Run("jwk parse request", func(t *testing.T) {
		public, private := auth.ES256()

		e, cookieName := email.Email("foobar@gmail.com"), `__g`

		o := auth.TokenOption{
			Issuer:     "github.com/rog-golang-buddies/rmx",
			Subject:    suid.NewUUID().String(),
			Expiration: time.Hour * 10,
			Claims:     map[string]any{"email": e.String()},
		}

		// rts
		rts, err := auth.Sign(private, &o)
		is.NoErr(err) // signing refresh token

		c := &http.Cookie{
			Path:     "/",
			Name:     cookieName,
			Value:    string(rts),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   24 * 7,
		}
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(c)

		_, err = jwt.Parse([]byte(c.Value), jwt.WithKey(jwa.ES256, public), jwt.WithValidate(true))
		is.NoErr(err) // parsing jwk page not found
	})
}
