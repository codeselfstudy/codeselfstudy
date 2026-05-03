package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// userContextKey is the echo.Context key under which validated claims are
// stored. Handlers reach for it via Claims(c).
const userContextKey = "auth:claims"

// Middleware returns Echo middleware that requires a valid WorkOS-signed JWT
// in the `Authorization: Bearer <token>` header. On success it places the
// parsed claims in the request context under userContextKey; on failure it
// returns 401 with no body. Signature, expiry, and issuer are all checked
// against the verifier's JWKS / configured issuer.
func Middleware(v *Verifier) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tok, err := authenticate(c, v)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}
			c.Set(userContextKey, tok)
			return next(c)
		}
	}
}

func authenticate(c echo.Context, v *Verifier) (jwt.Token, error) {
	raw := bearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
	if raw == "" {
		return nil, errors.New("auth: missing bearer token")
	}

	keys, err := v.keySet(c.Request().Context())
	if err != nil {
		return nil, err
	}

	tok, err := jwt.Parse(
		[]byte(raw),
		jwt.WithKeySet(keys),
		jwt.WithValidate(true),
		jwt.WithIssuer(v.issuer),
	)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

// Claims returns the validated JWT claims placed on the context by
// Middleware, or nil if the request didn't go through the middleware.
func Claims(c echo.Context) jwt.Token {
	v := c.Get(userContextKey)
	if v == nil {
		return nil
	}
	tok, _ := v.(jwt.Token)
	return tok
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return ""
	}
	return strings.TrimSpace(header[len(prefix):])
}
