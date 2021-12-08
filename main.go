package main

import (
	"log"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/internal/http"
	"github.com/p-l/fringe/internal/radius"
	"github.com/p-l/fringe/internal/repos"
	"github.com/spf13/viper"
	"modernc.org/ql"
)

func fatalOnInvalidConfig() {
	criticalKeys := []string{
		"fringe.allowed-domain",
		"fringe.secret",
		"radius.secret",
		"google-oauth.client-id",
		"google-oauth.client-secret",
	}

	for _, key := range criticalKeys {
		if len(viper.GetString(key)) == 0 {
			entries := strings.Split(key, ".")
			log.Panicf("FATAL: missing configuration key %s under section %s in configuration file", entries[1], entries[0])
		}
	}
}

func main() {
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("toml")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/fringe/") // path to look for the config file in
	viper.AddConfigPath(".")            // optionally look for config in the working directory

	// Default values
	viper.SetDefault("http.root", "http://127.0.0.1:9990/")
	viper.SetDefault("database.location", "/var/lib/fringe/users.repos")

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("Fatal error config file: %v", err)
	}

	// Validate configuration
	fatalOnInvalidConfig()

	// Initialize Database connexion
	ql.RegisterDriver()

	connexion, err := sqlx.Open("ql", viper.GetString("database.location"))
	if err != nil {
		log.Panicf("FATAL: Could not connect to database: %v", err)
	}
	defer func() { _ = connexion.Close() }() //nolint:wsl

	userRepo, err := repos.NewUserRepository(connexion)
	if err != nil {
		log.Panicf("FATAL: Could not initate user reposityr: %v", err)
	}

	go func() {
		radius.ServeRadius(userRepo, viper.GetString("radius.secret"))
	}()

	// HTTP
	http.ServeHTTP(
		userRepo,
		viper.GetString("http.root-url"),
		viper.GetString("google-oauth.client-id"),
		viper.GetString("google-oauth.client-secret"),
		viper.GetString("fringe.allowed-domain"),
		viper.GetString("fringe.secret"))
}
