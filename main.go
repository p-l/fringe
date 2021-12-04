package main

import (
	"log"
	"strings"

	"github.com/p-l/fringe/internal/db"
	"github.com/p-l/fringe/internal/http"
	"github.com/p-l/fringe/internal/radius"
	"github.com/spf13/viper"
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
	viper.SetDefault("database.location", "/var/lib/fringe/users.db")

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Fatal error config file: %v", err)
	}

	// Validate configuration
	fatalOnInvalidConfig()

	// Start
	dbLocation := viper.GetString("database.location")

	repo, err := db.NewRepository(dbLocation)
	if err != nil {
		log.Fatalf("Could not open or create database at %s: %v", dbLocation, err)
	}
	defer repo.Close()

	go func() {
		radius.ServeRadius(repo, viper.GetString("radius.secret"))
	}()

	// HTTP
	http.ServeHTTP(
		repo,
		viper.GetString("http.root-url"),
		viper.GetString("google-oauth.client-id"),
		viper.GetString("google-oauth.client-secret"),
		viper.GetString("fringe.allowed-domain"),
		viper.GetString("fringe.secret"))
}
