package system_test

import (
	"testing"

	"github.com/p-l/fringe/internal/system"
	"github.com/stretchr/testify/assert"
)

func TestLoadSecretsFromFile(t *testing.T) {
	t.Parallel()

	t.Run("Creates a new file when none not exist", func(t *testing.T) {
		t.Parallel()

		filename := t.TempDir() + "/secrets.json"
		secrets := system.LoadSecretsFromFile(filename)

		assert.FileExists(t, filename)
		assert.NotZero(t, len(secrets.JWT))
		assert.NotZero(t, len(secrets.Radius))
	})

	t.Run("Fills missing Radius secrets on loading", func(t *testing.T) {
		t.Parallel()

		filename := t.TempDir() + "/secrets.json"
		secrets := system.Secrets{
			Radius: "",
			JWT:    "1234567890",
		}
		system.SaveSecretsToFile(secrets, filename)
		assert.FileExists(t, filename)

		loadedSecrets := system.LoadSecretsFromFile(filename)
		assert.NotZero(t, len(loadedSecrets.JWT))
		assert.NotZero(t, len(loadedSecrets.Radius))
	})

	t.Run("Fills missing JWT secrets on loading", func(t *testing.T) {
		t.Parallel()

		filename := t.TempDir() + "/secrets.json"
		secrets := system.Secrets{
			Radius: "1234567890",
			JWT:    "",
		}
		system.SaveSecretsToFile(secrets, filename)
		assert.FileExists(t, filename)

		loadedSecrets := system.LoadSecretsFromFile(filename)
		assert.NotZero(t, len(loadedSecrets.JWT))
		assert.NotZero(t, len(loadedSecrets.Radius))
	})

	t.Run("Load secrets unchanged when file exists", func(t *testing.T) {
		t.Parallel()

		filename := t.TempDir() + "/secrets.json"
		secrets := system.Secrets{
			Radius: "radius",
			JWT:    "jwt",
		}
		system.SaveSecretsToFile(secrets, filename)
		assert.FileExists(t, filename)

		loadedSecrets := system.LoadSecretsFromFile(filename)
		assert.Equal(t, "radius", loadedSecrets.Radius)
		assert.Equal(t, "jwt", loadedSecrets.JWT)
	})
}
