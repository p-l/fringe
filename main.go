package main

import (
	"context"
	"crypto/tls"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/p-l/fringe/assets"
	"github.com/p-l/fringe/internal/httpd"
	"github.com/p-l/fringe/internal/radiusd"
	"github.com/p-l/fringe/internal/repos"
	"github.com/p-l/fringe/internal/system"
	"github.com/p-l/fringe/templates"
	"github.com/spf13/viper"
	"golang.org/x/crypto/acme/autocert"
	"layeh.com/radius"
	"modernc.org/ql"
)

const terminationWait = time.Second * 5

func openDB(databaseFile string) *sqlx.DB {
	// Initialize Database connexion
	ql.RegisterDriver()

	db, err := sqlx.Open("ql", databaseFile)
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

func newWebServers(config system.Config, userRepo *repos.UserRepository, jwtSecret string) (*http.Server, *http.Server) {
	staticTemplates := fs.FS(templates.Files())
	staticAssets := fs.FS(assets.Files())

	// HTTPS
	httpsSrv := httpd.NewHTTPServer(
		config,
		userRepo,
		staticTemplates,
		staticAssets,
		jwtSecret)

	// TLS Cert Manager
	var certManager *autocert.Manager

	var tlsConfig *tls.Config

	if config.Web.UseLetsEncrypt {
		certManager = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.Web.Domain),
			Cache:      autocert.DirCache("certs"),
		}
		tlsConfig = &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		}
	} else {
		tlsConfig = system.TLSConfigWithSelfSignedCert(system.AllLocalIPAddresses())
	}

	// Add the TLS configuration to the https server
	httpsSrv.TLSConfig = tlsConfig

	// HTTP to HTTPS Redirection and autocert server
	redirectSrv := httpd.NewRedirectServer(
		config.Services.HTTPBindAddress,
		config.Web.Domain,
		certManager)

	return httpsSrv, redirectSrv
}

func main() {
	// Load and validate configuration
	viperConf := viper.New()
	viperConf.SetConfigName("config")       // name of config file (without extension)
	viperConf.SetConfigType("toml")         // REQUIRED if the config file does not have the extension in the name
	viperConf.AddConfigPath("/etc/fringe/") // path to look for the config file in
	viperConf.AddConfigPath(".")            // optionally look for config in the working directory
	config := system.LoadConfig(viperConf)

	// Get the Secrets
	secrets := system.LoadSecretsFromFile(config.Storage.SecretsFile)

	// Get User Repository
	db := openDB(config.Storage.UserDatabaseFile)
	userRepo := openUserRepo(db)

	// Servers
	radiusSrv := radiusd.NewRadiusServer(userRepo, secrets.Radius)
	httpsSrv, redirectSrv := newWebServers(config, userRepo, secrets.JWT)

	// Start Radius
	go func() {
		if err := radiusSrv.ListenAndServe(); err != nil {
			log.Panicf("radius server died with error: %v", err)
		}
	}()

	// Start HTTPS
	go func() {
		if err := httpsSrv.ListenAndServeTLS("", ""); err != nil {
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
