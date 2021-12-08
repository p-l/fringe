package http

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/http/handlers"
	"github.com/p-l/fringe/internal/http/helpers"
	"github.com/p-l/fringe/internal/http/middlewares"
	"github.com/p-l/fringe/internal/repos"
)

//const (
//	generatedPasswordLen         = 24
//	jwtValidityDurationInMinutes = 5
//)

// ServeHTTP Starts blocking HTTP server.
func ServeHTTP(repo *repos.UserRepository, rootURL string, googleClientID string, googleClientSecret string, allowedDomain string, jwtSecret string) {
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

	log.Printf("Starting http server on :9990")

	err := http.ListenAndServe(":9990", router)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
}
