package main

import (
	"context"
	"crypto/tls"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/assets"
	"github.com/p-l/fringe/internal/httpd"
	"github.com/p-l/fringe/internal/radiusd"
	"github.com/p-l/fringe/internal/repos"
	"github.com/p-l/fringe/templates"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"layeh.com/radius"
	"modernc.org/ql"
)

const terminationWait = time.Second * 5

func loadConfigFile() {
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("toml")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/etc/fringe/") // path to look for the config file in
	viper.AddConfigPath(".")            // optionally look for config in the working directory

	// Default values
	viper.SetDefault("web.http-bind-address", ":80")
	viper.SetDefault("web.https-bind-address", ":443")
	viper.SetDefault("web.http-redirect", false)
	viper.SetDefault("web.domain", "127.0.0.1")
	viper.SetDefault("web.use-letsencrypt", true)
	viper.SetDefault("database.location", "/var/lib/fringe/users.repos")

	// Read the configuration
	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("config file error: %v", err)
	}
}

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
			log.Panicf("missing configuration key %s under section %s in configuration file", entries[1], entries[0])
		}
	}

	if !viper.GetBool("web.lets-encrypt") {
		tlsCertFile := viper.GetString("web.tls-cert-file")
		tlsKeyFile := viper.GetString("web.tls-key-file")

		if _, err := os.Stat(tlsCertFile); errors.Is(err, os.ErrNotExist) {
			log.Panicf("lets-encrypt is turned off and the tls-cert-file (%s) is not found", tlsCertFile)
		}

		if _, err := os.Stat(tlsKeyFile); errors.Is(err, os.ErrNotExist) {
			log.Panicf("lets-encrypt is turned off and the tls-key-file (%s) is not found", tlsKeyFile)
		}
	}
}

func openDB() *sqlx.DB {
	// Initialize Database connexion
	ql.RegisterDriver()

	db, err := sqlx.Open("ql", viper.GetString("database.location"))
	if err != nil {
		log.Panicf("could not connect to database: %v", err)
	}

	return db
}

func openUserRepo(connexion *sqlx.DB) *repos.UserRepository {
	userRepo, err := repos.NewUserRepository(connexion)
	if err != nil {
		log.Panicf("could not initate user repository: %v", err)
	}

	return userRepo
}

func httpSrvTextsFromConfig() httpd.Texts {
	return httpd.Texts{
		PasswordHint:          viper.GetString("texts.password.hint"),
		PasswordInfoCardTitle: viper.GetString("texts.password.info-title"),
		PasswordInfoCardItems: viper.GetStringSlice("texts.password.info-items"),
	}
}

func newWebServers(userRepo *repos.UserRepository) (*http.Server, *http.Server) {
	staticTemplates := fs.FS(templates.Files())
	staticAssets := fs.FS(assets.Files())
	httpSrvTexts := httpSrvTextsFromConfig()

	// HTTPS
	httpsSrv := httpd.NewHTTPServer(
		userRepo,
		staticTemplates,
		staticAssets,
		viper.GetString("web.https-bind-address"),
		viper.GetString("web.domain"),
		viper.GetStringSlice("fringe.admin-emails"),
		viper.GetString("google-oauth.client-id"),
		viper.GetString("google-oauth.client-secret"),
		viper.GetString("fringe.allowed-domain"),
		viper.GetString("fringe.secret"),
		httpSrvTexts)

	// TLS Cert Manager
	var certManager *autocert.Manager
	tlsConfig := tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	if viper.GetBool("web.lets-encrypt") {
		certManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(viper.GetString("web.domain")),
			Cache:      autocert.DirCache("certs"),
		}
		tlsConfig.GetCertificate = certManager.GetCertificate
	}

	// Add the TLS configuration to the https server
	httpsSrv.TLSConfig = &tlsConfig

	// HTTP to HTTPS Redirection and autocert server
	redirectSrv := httpd.NewRedirectServer(
		viper.GetString("web.http-bind-address"),
		viper.GetString("web.domain"),
		certManager)

	return httpsSrv, redirectSrv
}

func main() {
	// Load and validate configuration
	loadConfigFile()
	fatalOnInvalidConfig()

	db := openDB()
	userRepo := openUserRepo(db)

	// Servers
	radiusSrv := radiusd.NewRadiusServer(userRepo, viper.GetString("radius.secret"))
	httpsSrv, redirectSrv := newWebServers(userRepo)

	// Start Radius
	go func() {
		if err := radiusSrv.ListenAndServe(); err != nil {
			log.Panicf("radius server died with error: %v", err)
		}
	}()

	// Start HTTPS
	go func() {
		if err := httpsSrv.ListenAndServeTLS(viper.GetString("web.tls-cert-file"), viper.GetString("web.tls-key-file")); err != nil {
			log.Panicf("web server (https) died with error: %v", err)
		}
	}()

	// Start HTTP
	go func() {
		if err := redirectSrv.ListenAndServe(); err != nil {
			log.Panicf("web redirect server (http) died with error: %v", err)
		}
	}()

	waitOn(httpsSrv, redirectSrv, radiusSrv, db)
}

func waitOn(httpSrv *http.Server, redirectSrv *http.Server, radiusSrv *radius.PacketServer, connexion *sqlx.DB) {
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
	_ = redirectSrv.Shutdown(ctx)
	_ = connexion.Close()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0) //nolint:gocritic
}
