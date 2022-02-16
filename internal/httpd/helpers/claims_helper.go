package helpers

import (
	"context"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

const AuthClaimsDurationInMinutes = 60

type userCtxKeyType string

const userCtxKey userCtxKeyType = "auth_claims"

type AuthClaims struct {
	Email       string `json:"email"`
	Name        string `json:"name"`
	Picture     string `json:"picture"`
	Permissions string `json:"permissions"`
	jwt.StandardClaims
}

func NewAuthClaims(email string, name string, picture string, permissions string) *AuthClaims {
	expirationTime := time.Now().Add(AuthClaimsDurationInMinutes * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	return &AuthClaims{
		Email:       email,
		Name:        name,
		Picture:     picture,
		Permissions: permissions,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}
}

func AuthClaimsFromContext(ctx context.Context) (*AuthClaims, bool) {
	claims, ok := ctx.Value(userCtxKey).(*AuthClaims)
	if !ok {
		return nil, ok
	}

	return claims, ok
}

func (c *AuthClaims) ContextWithClaims(ctx context.Context) context.Context {
	return context.WithValue(ctx, userCtxKey, c)
}

func (c *AuthClaims) Refresh() *AuthClaims {
	expirationTime := time.Now().Add(AuthClaimsDurationInMinutes * time.Minute)
	c.StandardClaims.ExpiresAt = expirationTime.Unix()

	return c
}

func (c *AuthClaims) IsAdmin() bool {
	return strings.EqualFold(c.Permissions, AdminRoleString)
}
