package httpd

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/p-l/fringe/internal/httpd/services"
	"github.com/p-l/fringe/internal/repos"
)

type Texts struct {
	PasswordHint          string
	PasswordInfoCardTitle string
	PasswordInfoCardItems []string
}

const httpTimeouts = time.Second * 5

// NewHTTPServer Create and configure the HTTP server.
func NewHTTPServer(repo *repos.UserRepository, templates fs.FS, assets fs.FS, rootURL string, adminEmails []string, googleClientID string, googleClientSecret string, allowedDomain string, jwtSecret string, texts Texts) *http.Server {
	googleOAuth := services.NewGoogleOAuthService(http.DefaultClient, googleClientID, googleClientSecret, fmt.Sprintf("%s/auth/google/callback", rootURL))

	authHelper := helpers.NewAuthHelper(allowedDomain, jwtSecret, adminEmails)
	pageHelper := helpers.NewPageHelper(templates)

	logMiddleware := middlewares.NewLogMiddleware()
	authMiddleware := middlewares.NewAuthMiddleware("/auth/", []string{"/assets"}, authHelper)

	homeHandler := handlers.NewDefaultHandler(repo, pageHelper)
	authHandler := handlers.NewAuthHandler(googleOAuth, authHelper)
	userHandler := handlers.NewUserHandler(repo, authHelper, pageHelper, texts.PasswordHint, texts.PasswordInfoCardTitle, texts.PasswordInfoCardItems)

	router := mux.NewRouter()
	router.Use(logMiddleware.LogRequests)
	router.Use(authMiddleware.EnsureAuth)

	router.HandleFunc("/", homeHandler.Root).Methods(http.MethodGet)
	router.HandleFunc("/auth/", authHandler.RootHandler).Methods(http.MethodGet)
	router.HandleFunc("/auth/google/callback", authHandler.GoogleCallbackHandler).Methods(http.MethodGet)
	router.HandleFunc("/user/", userHandler.List).Methods("GET")
	router.HandleFunc("/user/{email}/", userHandler.View).Methods(http.MethodGet, http.MethodDelete)
	router.HandleFunc("/user/{email}/enroll", userHandler.Enroll).Methods(http.MethodGet)
	router.HandleFunc("/user/{email}/password", userHandler.Renew).Methods(http.MethodGet)
	router.HandleFunc("/user/{email}/delete", userHandler.Delete).Methods(http.MethodGet)
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	router.NotFoundHandler = http.HandlerFunc(homeHandler.NotFound)
	http.Handle("/", router)

	log.Printf("Created httpd server on :9990")
	httpd := http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:9990",
		WriteTimeout: httpTimeouts,
		ReadTimeout:  httpTimeouts,
		IdleTimeout:  httpTimeouts,
	}

	return &httpd
}
