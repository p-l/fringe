package helpers

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt"
)

const AuthClaimsDurationInMinutes = 5

type userCtxKeyType string

const userCtxKey userCtxKeyType = "auth_claims"

type AuthClaims struct {
	Email string `json:"email"`
	jwt.StandardClaims
}

func NewAuthClaims(email string) *AuthClaims {
	expirationTime := time.Now().Add(AuthClaimsDurationInMinutes * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	return &AuthClaims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}
}

func (c *AuthClaims) ContextWithClaims(ctx context.Context) context.Context {
	return context.WithValue(ctx, userCtxKey, c)
}

func AuthClaimsFromContext(ctx context.Context) (*AuthClaims, bool) {
	claims, ok := ctx.Value(userCtxKey).(*AuthClaims)
	if !ok {
		return nil, ok
	}

	return claims, ok
}

func (c *AuthClaims) Refresh() *AuthClaims {
	expirationTime := time.Now().Add(AuthClaimsDurationInMinutes * time.Minute)
	c.StandardClaims.ExpiresAt = expirationTime.Unix()

	return c
}
