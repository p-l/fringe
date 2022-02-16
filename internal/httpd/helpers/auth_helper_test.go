package helpers_test

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jaswdr/faker"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/stretchr/testify/assert"
)

func TestAuthHelper_NewJWTSignedString(t *testing.T) {
	t.Parallel()
	t.Run("returns single string without spaces", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		token := authHelper.NewJWTSignedString(claims)

		assert.NotContains(t, token, " ")
	})
}

func TestAuthHelper_AuthClaimsFromSignedToken(t *testing.T) {
	t.Parallel()
	t.Run("Return error on expired claims", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		// Force expiry to be 1 minute ago
		claims.StandardClaims.ExpiresAt = time.Now().Add(-1 * time.Minute).Unix()
		token := authHelper.NewJWTSignedString(claims)

		claimsFromToken, err := authHelper.AuthClaimsFromSignedToken(token)
		assert.Nil(t, claimsFromToken)
		assert.Error(t, err)
		assert.Equal(t, helpers.ErrInvalidClaimsToken, err)
	})

	t.Run("Return error on invalid value", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		k1Helper := helpers.NewAuthHelper("@test.com", "key1", []string{})
		k2Helper := helpers.NewAuthHelper("@test.com", "key2", []string{})

		k1claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		k1token := k1Helper.NewJWTSignedString(k1claims)

		claimsFromToken, err := k2Helper.AuthClaimsFromSignedToken(k1token)
		assert.Nil(t, claimsFromToken)
		assert.Error(t, err)
		assert.Equal(t, helpers.ErrInvalidClaimsToken, err)
	})

	t.Run("Refuses modified string", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		sourceToken := authHelper.NewJWTSignedString(claims)

		token := strings.ReplaceAll(sourceToken, ".", "&")
		token = strings.ReplaceAll(token, "_", ".")
		token = strings.ReplaceAll(token, "&", "_")

		assert.NotEqual(t, sourceToken, token)

		claimsFromToken, err := authHelper.AuthClaimsFromSignedToken(token)
		assert.Nil(t, claimsFromToken)
		assert.Error(t, err)
		assert.Equal(t, helpers.ErrInvalidClaimsToken, err)
	})

	t.Run("Provides claims from valid token", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()
		secret := fake.Internet().Password()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtKey := []byte(secret)

		tokenString, err := token.SignedString(jwtKey)
		assert.NoError(t, err)

		authHelper := helpers.NewAuthHelper("test.com", secret, []string{})

		claimsFromToken, err := authHelper.AuthClaimsFromSignedToken(tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, claimsFromToken)
		assert.Equal(t, claims.Email, claimsFromToken.Email)
		assert.Equal(t, claims.StandardClaims.ExpiresAt, claimsFromToken.StandardClaims.ExpiresAt)
	})
}

func TestAuthHelper_InAllowedDomain(t *testing.T) {
	t.Parallel()

	t.Run("True on email in domain", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		assert.True(t, authHelper.InAllowedDomain("email@test.com"))
	})

	t.Run("False on email outside of domain", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		assert.False(t, authHelper.InAllowedDomain("email@not-test.com"))
	})

	t.Run("False on invalid emails", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		assert.False(t, authHelper.InAllowedDomain("email@not-email"))
		assert.False(t, authHelper.InAllowedDomain("rabbit"))
		assert.False(t, authHelper.InAllowedDomain("(ihtrei485^%#)(2!~$"))
	})
}

func TestAuthHelper_RoleForEmail(t *testing.T) {
	t.Parallel()

	t.Run("admin for email in admin list", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{"admin@test.com"})
		assert.Equal(t, authHelper.RoleForEmail("admin@test.com"), helpers.AdminRoleString)
		assert.Equal(t, authHelper.RoleForEmail("ADMIN@test.com"), helpers.AdminRoleString)
		assert.Equal(t, authHelper.RoleForEmail("admin@TEST.com"), helpers.AdminRoleString)
	})

	t.Run("user for email not in admin list", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{"admin@test.com"})
		role := authHelper.RoleForEmail("user@test.com")
		assert.Equal(t, role, helpers.UserRoleString)
	})

	t.Run("False on email outside of domain", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{"admin@test.com"})
		assert.False(t, authHelper.InAllowedDomain("email@not-test.com"))
	})
}
