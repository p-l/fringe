package main

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/internal/httpd"
	"github.com/p-l/fringe/internal/radiusd"
	"github.com/p-l/fringe/internal/repos"
	"github.com/spf13/viper"
	"layeh.com/radius"
	"log"
	"modernc.org/ql"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
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

const terminationWait = time.Duration(time.Second * 5)

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

	userRepo, err := repos.NewUserRepository(connexion)
	if err != nil {
		log.Panicf("FATAL: Could not initate user reposityr: %v", err)
	}

	// Get Radius Started
	radiusSrv := radiusd.NewRadiusServer(userRepo, viper.GetString("radius.secret"))
	go func() {
		if err := radiusSrv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// HTTP
	httpSrv := httpd.NewHTTPServer(
		userRepo,
		viper.GetString("http.root-url"),
		viper.GetString("google-oauth.client-id"),
		viper.GetString("google-oauth.client-secret"),
		viper.GetString("fringe.allowed-domain"),
		viper.GetString("fringe.secret"))
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	waitOn(httpSrv, radiusSrv, connexion)
}

func waitOn(httpSrv *http.Server, radiusSrv *radius.PacketServer, connexion *sqlx.DB) {
	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT or SIGTERM
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.

	ctx, cancel := context.WithTimeout(context.Background(), terminationWait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	_ = httpSrv.Shutdown(ctx)
	_ = radiusSrv.Shutdown(ctx)
	_ = connexion.Close()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}
