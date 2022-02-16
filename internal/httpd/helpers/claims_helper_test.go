package helpers_test

import (
	"context"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthClaims(t *testing.T) {
	t.Parallel()

	t.Run("new claims expire in the future", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), helpers.UserRoleString)

		assert.Greater(t, claims.StandardClaims.ExpiresAt, time.Now().Unix())
	})
}

func TestAuthClaimsFromContext(t *testing.T) {
	t.Parallel()

	t.Run("returns failure if no claims are present", func(t *testing.T) {
		t.Parallel()

		claimsFromContext, ok := helpers.AuthClaimsFromContext(context.Background())

		assert.False(t, ok)
		assert.Nil(t, claimsFromContext)
	})

	t.Run("returns unmodified claims from the context", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), helpers.UserRoleString)
		claimsCtx := claims.ContextWithClaims(context.Background())

		claimsFromContext, ok := helpers.AuthClaimsFromContext(claimsCtx)
		assert.True(t, ok)
		assert.NotNil(t, claimsFromContext)
		assert.Equal(t, claimsFromContext.Email, claims.Email)
		assert.Equal(t, claimsFromContext.StandardClaims.ExpiresAt, claims.StandardClaims.ExpiresAt)
	})
}

func TestAuthClaims_ContextWithClaims(t *testing.T) {
	t.Parallel()

	t.Run("Stores claims in context without modification", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), helpers.UserRoleString)
		claimsCtx := claims.ContextWithClaims(context.Background())

		claimsFromContext, ok := helpers.AuthClaimsFromContext(claimsCtx)
		assert.True(t, ok)
		assert.NotNil(t, claimsFromContext)
		assert.Equal(t, claims.Email, claimsFromContext.Email)
		assert.Equal(t, claims.StandardClaims.ExpiresAt, claimsFromContext.StandardClaims.ExpiresAt)
	})
}

func TestAuthClaims_Refresh(t *testing.T) {
	t.Parallel()

	t.Run("Updates expiry on struct", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), helpers.UserRoleString)
		claimsOriginalExpiry := time.Now().Add(-1 * time.Minute).Unix()
		claims.StandardClaims.ExpiresAt = claimsOriginalExpiry
		returnedClaims := claims.Refresh()

		assert.NotEqual(t, claimsOriginalExpiry, claims.StandardClaims.ExpiresAt)
		assert.Greater(t, claims.StandardClaims.ExpiresAt, time.Now().Unix())
		assert.Equal(t, claims, returnedClaims)
	})
}

func TestAuthClaims_IsAdmin(t *testing.T) {
	t.Parallel()

	t.Run("Return true if email is has the admin role", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), helpers.AdminRoleString)

		assert.True(t, claims.IsAdmin())
	})

	t.Run("Return false if email is has the user role", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), helpers.UserRoleString)

		assert.False(t, claims.IsAdmin())
	})

	t.Run("Return false if email is has empty role", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")

		assert.False(t, claims.IsAdmin())
	})

	t.Run("Return false if email is has unknown role", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "unknown")

		assert.False(t, claims.IsAdmin())
	})
}
