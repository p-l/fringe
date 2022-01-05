package system

import (
	"log"
	"net"

	"github.com/spf13/viper"
)

type SecurityConfig struct {
	AllowedDomain         string   `mapstructure:"allowed-domain"`
	AuthorizedAdminEmails []string `mapstructure:"admin-emails"` //nolint:tagliatelle
}

type WebConfig struct {
	Domain         string `mapstructure:"domain"`
	UseLetsEncrypt bool   `mapstructure:"lets-encrypt"` //nolint:tagliatelle
}

type StorageConfig struct {
	UserDatabaseFile string `mapstructure:"user-database"` //nolint:tagliatelle
	SecretsFile      string `mapstructure:"secrets-file"`
}

type ServicesConfig struct {
	HTTPBindAddress   string `mapstructure:"http-bind-address"`
	HTTPSBindAddress  string `mapstructure:"https-bind-address"`
	RadiusBindAddress string `mapstructure:"radius-bind-address"`
}

type GoogleConfig struct {
	ClientID     string `mapstructure:"client-id"`
	ClientSecret string `mapstructure:"client-secret"`
}

type OAuthConfig struct {
	Google GoogleConfig `mapstructure:"google"`
}

type Config struct {
	OAuth    OAuthConfig    `mapstructure:"oauth"` //nolint:tagliatelle
	Security SecurityConfig `mapstructure:"security"`
	Services ServicesConfig `mapstructure:"services"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Web      WebConfig      `mapstructure:"web"`
}

func LoadConfig() Config {
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("toml")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/fringe/") // path to look for the config file in
	viper.AddConfigPath(".")            // optionally look for config in the working directory

	// Default values
	viper.SetDefault("services.https-bind-address", ":443")
	viper.SetDefault("services.http-bind-address", ":80")
	viper.SetDefault("services.radius-bind-address", ":1812")

	localIP := FirstLocalIP(AllLocalIPAddresses()).String()
	viper.SetDefault("web.domain", localIP)
	viper.SetDefault("web.lets-encrypt", true)
	viper.SetDefault("storage.user-database", "/var/lib/fringe/users.repos")
	viper.SetDefault("storage.secrets-file", "/var/lib/fringe/secrets.json")

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("config file error: %v", err)
	}

	// Disable lets-encrypt if domain is an IP
	domain := viper.GetString("web.domain")

	if addr := net.ParseIP(domain); addr != nil {
		viper.Set("web.lets-encrypt", false)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Panicf("could not parse configuration: %v", err)
	}

	return config
}
