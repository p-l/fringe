package main

import (
	"log"

	"github.com/dchest/uniuri"
	_ "github.com/mattn/go-sqlite3"
	"github.com/p-l/fringe/internal/db"
	"github.com/p-l/fringe/internal/http"
	"github.com/p-l/fringe/internal/radius"
	"github.com/spf13/viper"
)

const randomRadiusSecretLen = 16

func main() {
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("toml")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/fringe/") // path to look for the config file in
	viper.AddConfigPath(".")            // optionally look for config in the working directory

	// Default values
	randomString := uniuri.NewLen(randomRadiusSecretLen)
	viper.SetDefault("radius.secret", randomString)
	viper.SetDefault("google-oauth.client-id", "set-client-id-in-config.toml")
	viper.SetDefault("http.root", "http://127.0.0.1:9990/")

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Fatal error config file: %v", err)
	}

	// Load resulting values
	radiusSecret := viper.GetString("radius.secret")
	googleClientID := viper.GetString("google-oauth.client-id")
	googleClientSecret := viper.GetString("google-oauth.client-secret")
	httpRootURL := viper.GetString("http.root-url")
	allowedDomains := viper.GetString("fringe.allowed-domain")
	jwtSecret := viper.GetString("fringe.secret")

	repo, _ := db.NewRepository()
	defer repo.Close()

	go func() {
		if radiusSecret == randomString {
			log.Printf("INFO: radius.secret not set in configuration using random secret: %s", radiusSecret)
		}

		radius.ServeRadius(repo, radiusSecret)
	}()

	// HTTP
	http.ServeHTTP(repo, httpRootURL, googleClientID, googleClientSecret, allowedDomains, jwtSecret)
}
