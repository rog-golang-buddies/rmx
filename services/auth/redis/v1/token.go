package token

import (
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/rog-golang-buddies/rmx/pkg/auth"
)

type Client struct {
	algorithm jwa.SignatureAlgorithm
	TokenExp  map[string]time.Duration
}

// GenerateToken implements Client
func (*Client) GenerateToken(private jwk.Key, opts *auth.TokenOption) ([]byte, error) {
	return auth.Sign(private, opts)
}

// GenerateKeys implements Client
//
// If a symmetrical algorithm is selected, then private and public keys will be the same
func (c *Client) GenerateKeys() (public jwk.Key, private jwk.Key, err error) {
	switch algo := c.algorithm; algo {
	case jwa.HS256, jwa.HS384, jwa.HS512:
		return nil, nil, fmt.Errorf("symmetrical algorithms are not supported")
	case jwa.RS256, jwa.RS384, jwa.RS512:
		pb, pr := auth.RS256()
		return pb, pr, nil
	case jwa.ES256, jwa.ES384, jwa.ES512:
		pb, pr := auth.ES256()
		return pb, pr, nil
	}

	return nil, nil, fmt.Errorf("unsupported algorithm: %s", c.algorithm)
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		TokenExp: make(map[string]time.Duration),
	}

	for _, opt := range opts {
		opt(c)
	}

	// Set default algorithm
	if c.algorithm == "" {
		// NOTE - default to a strong asymmetric algorithm
		c.algorithm = jwa.RS256
	}

	// Set default token expirations
	if _, ok := c.TokenExp["id"]; !ok {
		c.TokenExp["id"] = time.Hour * 10
	}

	if _, ok := c.TokenExp["access"]; !ok {
		c.TokenExp["access"] = time.Hour * 24 * 7
	}

	if _, ok := c.TokenExp["refresh"]; !ok {
		c.TokenExp["refresh"] = time.Minute * 5
	}

	return c
}

type ClientOption func(*Client)

func WithAlgorithm(algorithm jwa.SignatureAlgorithm) ClientOption {
	return func(c *Client) {
		c.algorithm = algorithm
	}
}

// NOTE - make this variadic
func WithTokenTimeout(idTokenExp, accessTokenExp, refreshTokenExp time.Duration) ClientOption {
	// map[string]time.Duration{}
	return func(s *Client) {
		s.TokenExp["id"] = idTokenExp
		s.TokenExp["access"] = accessTokenExp
		s.TokenExp["refresh"] = refreshTokenExp
	}
}
