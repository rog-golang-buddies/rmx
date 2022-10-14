package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
	"github.com/rog-golang-buddies/rmx/internal"
	"github.com/rog-golang-buddies/rmx/internal/fp"
)

type Client struct {
	rtdb, cidb *redis.Client
}

var (
	ErrNotImplemented = errors.New("not implemented")
	ErrGenerateKey    = errors.New("failed to generate new ecdsa key pair")
	ErrSignTokens     = errors.New("failed to generate signed tokens")
	ErrRTValidate     = errors.New("failed to validate refresh token")
)

func NewRedis(addr, password string) *Client {
	rtdb := redis.Options{Addr: addr, Password: password, DB: 0}
	cidb := redis.Options{Addr: addr, Password: password, DB: 1}

	c := &Client{redis.NewClient(&rtdb), redis.NewClient(&cidb)}
	return c
}

const (
	defaultAddr     = "localhost:6379"
	defaultPassword = ""
)

var DefaultClient = &Client{
	rtdb: redis.NewClient(&redis.Options{Addr: defaultAddr, Password: defaultPassword, DB: 0}),
	cidb: redis.NewClient(&redis.Options{Addr: defaultAddr, Password: defaultPassword, DB: 1}),
}

func (c *Client) ValidateRefreshToken(ctx context.Context, token string) error {
	tc, err := ParseRefreshTokenClaims(token)
	if err != nil {
		return err
	}

	cid := tc.Subject()
	email, ok := tc.PrivateClaims()["email"].(string)
	if !ok {
		return ErrRTValidate
	}

	if err := c.ValidateClientID(ctx, cid); err != nil {
		return err
	}

	if _, err := c.rtdb.Get(ctx, token).Result(); err != nil {
		switch err {
		case redis.Nil:
			return nil
		default:
			return err
		}
	}

	err = c.BlackListClientID(ctx, cid, email)
	if err != nil {
		return err
	}

	return ErrRTValidate
}

func (c *Client) BlackListClientID(ctx context.Context, cid, email string) error {
	_, err := c.cidb.Set(ctx, cid, email, RefreshTokenExpiry).Result()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) BlackListRefreshToken(ctx context.Context, token string) error {
	_, err := c.rtdb.Set(ctx, token, nil, RefreshTokenExpiry).Result()
	return err
}

func (c *Client) ValidateClientID(ctx context.Context, cid string) error {
	// check if a key with client id exists
	// if the key exists it means that the client id is revoked and token should be denied
	// we don't need the email value here
	_, err := c.cidb.Get(ctx, cid).Result()
	if err != nil {
		switch err {
		case redis.Nil:
			return nil
		default:
			return ErrRTValidate
		}
	}

	return ErrRTValidate
}

// Easier to pass an array that two variables with context
type Pair [2]jwk.Key

func (p *Pair) Private() jwk.Key { return p[0] }
func (p *Pair) Public() jwk.Key  { return p[1] }

func NewPairES256() Pair {
	rawPriv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	key, err := jwk.FromRaw(rawPriv)
	if err != nil {
		panic(err)
	}

	_, ok := key.(jwk.ECDSAPrivateKey)
	if !ok {
		panic(ErrGenerateKey)
	}

	pub, err := key.PublicKey()
	if err != nil {
		panic(err)
	}

	return Pair{key, pub}
}

func NewPairRS256() Pair {
	rawPrv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	jwkPrv, err := jwk.FromRaw(rawPrv)
	if err != nil {
		panic(err)
	}

	jwkPub, err := jwkPrv.PublicKey()
	if err != nil {
		panic(err)
	}

	return Pair{jwkPrv, jwkPub}
}

func SignToken(key jwk.Key, opt *TokenOption) ([]byte, error) {
	var t time.Time
	if opt.IssuedAt.IsZero() {
		t = time.Now().UTC()
	} else {
		t = opt.IssuedAt
	}

	token, err := jwt.NewBuilder().
		Issuer(opt.Issuer).
		Audience(opt.Audience).
		Subject(opt.Subject).
		IssuedAt(t).
		Expiration(t.Add(opt.Expiration)).
		Build()
	if err != nil {
		return nil, ErrSignTokens
	}

	for _, c := range opt.Claims {
		if !c.HasValue() {
			return nil, fp.ErrTuple
		}

		err := token.Set(c[0], c[1])
		if err != nil {
			return nil, ErrSignTokens
		}
	}

	var algo jwa.SignatureAlgorithm
	if opt.Algo == "" {
		algo = jwa.RS256
	} else {
		algo = opt.Algo
	}

	return jwt.Sign(token, jwt.WithKey(algo, key))
}

func ParseRefreshTokenClaims(token string) (jwt.Token, error) { return jwt.Parse([]byte(token)) }

func ParseRefreshTokenWithValidate(key *jwk.Key, token string) (jwt.Token, error) {
	payload, err := jwt.Parse([]byte(token),
		jwt.WithKey(jwa.ES256, key),
		jwt.WithValidate(true))
	if err != nil {
		return nil, err
	}

	return payload, nil
}

type TokenOption struct {
	Issuer     string
	Audience   []string
	Subject    string
	Claims     []fp.Tuple
	IssuedAt   time.Time
	Expiration time.Duration
	Algo       jwa.SignatureAlgorithm
}

type authCtxKey string

const (
	// RefreshTokenCookieName = "RMX_REFRESH_TOKEN"
	RefreshTokenExpiry = time.Hour * 24 * 7
	AccessTokenExpiry  = time.Minute * 5
	EmailKey           = authCtxKey("rmx-email")
)

/*
 */
func ParseAuth(
	algo jwa.SignatureAlgorithm,
	key jwk.Key,
	cookieName ...string,
) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		var at http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			token, err := jwt.ParseRequest(r, jwt.WithKey(algo, key))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			email, ok := token.PrivateClaims()["email"].(string)
			if !ok {
				// NOTE unsure if we need to write anything more to the body
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// NOTE convert email from `string` type to `internal.Email` ?
			r = r.WithContext(
				context.WithValue(r.Context(), internal.EmailKey, internal.Email(email)),
			)
			h.ServeHTTP(w, r)
		}

		var rt http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
			rc, err := r.Cookie(cookieName[0])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(
				[]byte(rc.Value),
				jwt.WithKey(algo, key),
				jwt.WithValidate(true),
			)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			email, ok := token.PrivateClaims()["email"].(string)
			if !ok {
				// NOTE unsure if we need to write anything more to the body
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), internal.EmailKey, internal.Email(email))
			// ctx = context.WithValue(ctx, internal.EmailKey, r)
			r = r.WithContext(context.WithValue(ctx, internal.TokenKey, token))
			h.ServeHTTP(w, r)
		}

		if len(cookieName) != 0 {
			return http.HandlerFunc(rt)
		}

		return http.HandlerFunc(at)
	}
}