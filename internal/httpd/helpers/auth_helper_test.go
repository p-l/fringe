package helpers_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jaswdr/faker"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/stretchr/testify/assert"
)

func TestAuthHelper_NewJWTCookieFromClaims(t *testing.T) {
	t.Parallel()
	t.Run("return cookie that expires with claims", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), "")
		cookie := authHelper.NewJWTCookieFromClaims(claims)

		assert.Equal(t, claims.ExpiresAt, cookie.Expires.Unix())
	})

	t.Run("returns a secure cookie ", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), "")
		cookie := authHelper.NewJWTCookieFromClaims(claims)

		assert.Equal(t, "token", cookie.Name)
		assert.True(t, cookie.Secure)
		assert.True(t, cookie.HttpOnly)
	})
}

func TestAuthHelper_AuthClaimsFromSignedToken(t *testing.T) {
	t.Parallel()
	t.Run("Return error on expired claims", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), "")
		// Force expiry to be 1 minute ago
		claims.StandardClaims.ExpiresAt = time.Now().Add(-1 * time.Minute).Unix()
		cookie := authHelper.NewJWTCookieFromClaims(claims)

		claimsFromToken, err := authHelper.AuthClaimsFromSignedToken(cookie.Value)
		assert.Nil(t, claimsFromToken)
		assert.Error(t, err)
		assert.Equal(t, helpers.ErrInvalidClaimsToken, err)
	})

	t.Run("Return error on invalid value", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		k1Helper := helpers.NewAuthHelper("@test.com", "key1", []string{})
		k2Helper := helpers.NewAuthHelper("@test.com", "key2", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), "")
		k1Cookie := k1Helper.NewJWTCookieFromClaims(claims)

		claimsFromToken, err := k2Helper.AuthClaimsFromSignedToken(k1Cookie.Value)
		assert.Nil(t, claimsFromToken)
		assert.Error(t, err)
		assert.Equal(t, helpers.ErrInvalidClaimsToken, err)
	})

	t.Run("Provides claims from valid token", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()
		secret := fake.Internet().Password()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), "")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtKey := []byte(secret)

		tokenString, err := token.SignedString(jwtKey)
		assert.NoError(t, err)

		authHelper := helpers.NewAuthHelper("@test.com", secret, []string{})

		claimsFromToken, err := authHelper.AuthClaimsFromSignedToken(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, claimsFromToken)
		assert.Equal(t, claims.Email, claimsFromToken.Email)
		assert.Equal(t, claims.StandardClaims.ExpiresAt, claimsFromToken.StandardClaims.ExpiresAt)
	})
}

func TestAuthHelper_RemoveJWTCookie(t *testing.T) {
	t.Parallel()
	t.Run("Return an expired cookie", func(t *testing.T) {
		t.Parallel()

		helper := helpers.NewAuthHelper("@test.com", "secret", []string{})

		cookie := helper.RemoveJWTCookie()

		assert.Equal(t, "token", cookie.Name)
		assert.Equal(t, cookie.Expires.Unix(), time.Unix(0, 0).Unix())
		assert.True(t, cookie.Secure)
		assert.True(t, cookie.HttpOnly)
	})
}
