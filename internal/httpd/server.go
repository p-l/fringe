package httpd

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/p-l/fringe/internal/repos"
)

//const (
//	generatedPasswordLen         = 24
//	jwtValidityDurationInMinutes = 5
//)

// ServeHTTP Starts blocking HTTP server.
func NewHTTPServer(repo *repos.UserRepository, rootURL string, googleClientID string, googleClientSecret string, allowedDomain string, jwtSecret string) *http.Server {
	googleOAuthConfig := handlers.GoogleOAuthClientConfig{
		ID:          googleClientID,
		Secret:      googleClientSecret,
		RedirectURL: fmt.Sprintf("%s/auth/google/callback", rootURL),
	}

	authHelper := helpers.NewAuthHelper(jwtSecret)

	logMiddleware := middlewares.NewLogMiddleware()
	authMiddleware := middlewares.NewAuthMiddleware("/auth", []string{"/assets"}, authHelper)

	homeHandler := handlers.NewHomeHandler(repo)
	authHandler := handlers.NewAuthHandler(allowedDomain, googleOAuthConfig, authHelper)
	userHandler := handlers.NewUserHandler(repo)

	router := mux.NewRouter()
	router.Use(logMiddleware.LogRequests)
	router.Use(authMiddleware.EnsureAuth)

	router.HandleFunc("/", homeHandler.ServeHome).Methods("GET")
	router.HandleFunc("/auth", authHandler.RootHandler).Methods("GET")
	router.HandleFunc("/auth/google/callback", authHandler.GoogleCallbackHandler).Methods("GET")
	router.HandleFunc("/user", userHandler.ServerIndex)
	router.HandleFunc("/user/new", userHandler.ServeNew)
	router.HandleFunc("/user/{email}", userHandler.ServeView)
	router.HandleFunc("/user/{email}/delete", userHandler.ServeDelete)
	// TODO: make asset location configurable?
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/"))))
	http.Handle("/", router)

	log.Printf("Created httpd server on :9990")
	httpd := http.Server{
		Handler:      router,
		Addr:         ":9990",
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	return &httpd
}
