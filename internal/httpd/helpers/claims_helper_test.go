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

		claims := helpers.NewAuthClaims(fake.Internet().Email())

		assert.Greater(t, claims.StandardClaims.ExpiresAt, time.Now().Unix())
	})
}

func TestContextWithClaims(t *testing.T) {
	t.Parallel()

	t.Run("Stores claims in context without modification", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email())
		claimsCtx := claims.ContextWithClaims(context.Background())

		claimsFromContext, ok := helpers.AuthClaimsFromContext(claimsCtx)
		assert.True(t, ok)
		assert.NotNil(t, claimsFromContext)
		assert.Equal(t, claims.Email, claimsFromContext.Email)
		assert.Equal(t, claims.StandardClaims.ExpiresAt, claimsFromContext.StandardClaims.ExpiresAt)
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

		claims := helpers.NewAuthClaims(fake.Internet().Email())
		claimsCtx := claims.ContextWithClaims(context.Background())

		claimsFromContext, ok := helpers.AuthClaimsFromContext(claimsCtx)
		assert.True(t, ok)
		assert.NotNil(t, claimsFromContext)
		assert.Equal(t, claimsFromContext.Email, claims.Email)
		assert.Equal(t, claimsFromContext.StandardClaims.ExpiresAt, claims.StandardClaims.ExpiresAt)
	})
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	t.Run("Updates expiry on struct", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		claims := helpers.NewAuthClaims(fake.Internet().Email())
		claimsOriginalExpiry := time.Now().Add(-1 * time.Minute).Unix()
		claims.StandardClaims.ExpiresAt = claimsOriginalExpiry
		returnedClaims := claims.Refresh()

		assert.NotEqual(t, claimsOriginalExpiry, claims.StandardClaims.ExpiresAt)
		assert.Greater(t, claims.StandardClaims.ExpiresAt, time.Now().Unix())
		assert.Equal(t, claims, returnedClaims)
	})
}
