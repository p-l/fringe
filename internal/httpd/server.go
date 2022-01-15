package httpd

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/p-l/fringe/internal/httpd/services"
	"github.com/p-l/fringe/internal/repos"
	"github.com/p-l/fringe/internal/system"
	"github.com/rs/cors"
)

type Texts struct {
	PasswordHint          string
	PasswordInfoCardTitle string
	PasswordInfoCardItems []string
}

const (
	httpsTimeouts        = time.Second * 5
	preFlightCacheMaxAge = time.Minute * 5
)

// NewHTTPServer Create and configure the HTTP server.
func NewHTTPServer(config system.Config, repo *repos.UserRepository, templates fs.FS, clientAssets fs.FS, jwtSecret string) *http.Server {
	googleOAuth := services.NewGoogleOAuthService(http.DefaultClient, config.OAuth.Google.ClientID, config.OAuth.Google.ClientSecret, fmt.Sprintf("https://%s/auth/google/callback", config.Web.Domain))

	authHelper := helpers.NewAuthHelper(config.Security.AllowedDomain, jwtSecret, config.Security.AuthorizedAdminEmails)
	pageHelper := helpers.NewPageHelper(templates)

	logMiddleware := middlewares.NewLogMiddleware()
	authMiddleware := middlewares.NewAuthMiddleware("/auth/", []string{"/api"}, []string{"/api/auth/", "/api/config/"}, authHelper)

	homeHandler := handlers.NewDefaultHandler(repo, pageHelper)
	authHandler := handlers.NewAuthHandler(googleOAuth, authHelper)
	userHandler := handlers.NewUserHandler(repo, authHelper, pageHelper)
	configHandler := handlers.NewConfigHandler(config.OAuth.Google)

	router := mux.NewRouter()
	router.Use(logMiddleware.LogRequests)
	router.Use(authMiddleware.EnsureAuth)

	// Hook the handlers
	router.HandleFunc("/api/", homeHandler.Root).Methods(http.MethodGet)
	router.HandleFunc("/api/auth/", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/config/", configHandler.Root).Methods(http.MethodGet)
	router.HandleFunc("/api/user/", userHandler.List).Methods("GET")
	router.HandleFunc("/api/user/{email}/", userHandler.View).Methods(http.MethodGet, http.MethodDelete)
	router.HandleFunc("/api/user/{email}/enroll", userHandler.Enroll).Methods(http.MethodGet)
	router.HandleFunc("/api/user/{email}/password", userHandler.Renew).Methods(http.MethodGet)
	router.HandleFunc("/api/user/{email}/delete", userHandler.Delete).Methods(http.MethodGet)

	// Serve the web client
	if len(config.Web.ReverseProxy) == 0 {
		// Built-in client
		router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.FS(clientAssets))))
		router.NotFoundHandler = http.HandlerFunc(homeHandler.NotFound)
	} else {
		log.Printf("Using reverse proxy from %s instead of serving files", config.Web.ReverseProxy)
		proxyURL, err := url.Parse(config.Web.ReverseProxy)
		if err != nil {
			log.Panicf("invalid reverse proxy url: %v", err)
		}

		reverseProxy := httputil.NewSingleHostReverseProxy(proxyURL)
		router.Handle("/", reverseProxy)
		router.NotFoundHandler = reverseProxy
	}

	httpdHandler := addCORS(config.Web.AllowOrigins, router)

	log.Printf("Created httpd server on %s", config.Services.HTTPSBindAddress)
	httpd := http.Server{
		Handler:      httpdHandler,
		Addr:         config.Services.HTTPSBindAddress,
		WriteTimeout: httpsTimeouts,
		ReadTimeout:  httpsTimeouts,
		IdleTimeout:  httpsTimeouts,
	}

	return &httpd
}

func addCORS(allowedOrigins []string, router *mux.Router) http.Handler {
	for _, origin := range allowedOrigins {
		log.Printf("HTTPD Allowed Origin: %s", origin)
	}

	corsHandler := cors.New(cors.Options{
		AllowCredentials: false,
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowedHeaders:   []string{"accept", "authorization", "content-type"},
		MaxAge:           int(preFlightCacheMaxAge.Seconds()),
		Debug:            true,
	})

	return corsHandler.Handler(router)
}
