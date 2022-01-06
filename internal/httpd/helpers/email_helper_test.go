package helpers_test

import (
	"testing"

	"github.com/jaswdr/faker"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/stretchr/testify/assert"
)

func TestIsEmailInDomain(t *testing.T) {
	t.Parallel()

	t.Run("Returns true if email is in specified domain", func(t *testing.T) {
		t.Parallel()

		assert.True(t, helpers.IsEmailInDomain("test@test.com", "test.com"))
	})

	t.Run("Returns false if email if specified domain includes a @", func(t *testing.T) {
		t.Parallel()

		assert.False(t, helpers.IsEmailInDomain("test@test.com", "@test.com"))
	})

	t.Run("Returns false if email if not in specified domain", func(t *testing.T) {
		t.Parallel()

		assert.False(t, helpers.IsEmailInDomain("test@test.com", "company.com"))
	})
}

func TestIsEmailValid(t *testing.T) {
	t.Parallel()

	t.Run("Returns true on valid emails", func(t *testing.T) {
		t.Parallel()

		fake := faker.New()

		assert.True(t, helpers.IsEmailValid(fake.Internet().Email()))
		assert.True(t, helpers.IsEmailValid(fake.Internet().FreeEmail()))
	})

	t.Run("Returns false on invalid emails", func(t *testing.T) {
		t.Parallel()

		fake := faker.New()

		assert.False(t, helpers.IsEmailValid(fake.Internet().User()))
		assert.False(t, helpers.IsEmailValid("invalid@email"))
		assert.False(t, helpers.IsEmailValid("@domain.com"))
	})
}
