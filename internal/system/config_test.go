package system_test

import (
	"os"
	"testing"

	"github.com/p-l/fringe/internal/system"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func newMockViperConfig(t *testing.T) *viper.Viper {
	t.Helper()

	tempDir := t.TempDir()
	viperConf := viper.New()
	viperConf.SetConfigName("config")
	viperConf.SetConfigType("toml")
	viperConf.AddConfigPath(tempDir)

	tempFile := tempDir + "/config.toml"
	file, _ := os.Create(tempFile)
	_ = file.Close()

	return viperConf
}

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	t.Run("Disable lets-encrypt if domain is an IP", func(t *testing.T) {
		t.Parallel()

		viperConf := newMockViperConfig(t)
		viperConf.Set("web.lets-encrypt", true)
		viperConf.Set("web.domain", "127.0.0.1")

		config := system.LoadConfig(viperConf)

		assert.False(t, config.Web.UseLetsEncrypt)
	})

	t.Run("Accept empty config", func(t *testing.T) {
		t.Parallel()

		viperConf := newMockViperConfig(t)
		config := system.LoadConfig(viperConf)

		assert.NotEmpty(t, config.Web.Domain)
		assert.False(t, config.Web.UseLetsEncrypt)
		assert.NotEmpty(t, config.Services.HTTPBindAddress)
		assert.NotEmpty(t, config.Services.HTTPSBindAddress)
		assert.NotEmpty(t, config.Services.RadiusBindAddress)
	})
}
