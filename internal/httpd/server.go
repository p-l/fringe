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
func NewHTTPServer(repo *repos.UserRepository, templates fs.FS, assets fs.FS, rootURL string, googleClientID string, googleClientSecret string, allowedDomain string, jwtSecret string, texts Texts) *http.Server {
	googleOAuth := services.NewGoogleOAuthService(http.DefaultClient, googleClientID, googleClientSecret, fmt.Sprintf("%s/auth/google/callback", rootURL))

	authHelper := helpers.NewAuthHelper(jwtSecret)
	pageHelper := helpers.NewPageHelper(templates)

	logMiddleware := middlewares.NewLogMiddleware()
	authMiddleware := middlewares.NewAuthMiddleware("/auth", []string{"/assets"}, authHelper)

	homeHandler := handlers.NewHomeHandler(repo)
	authHandler := handlers.NewAuthHandler(allowedDomain, googleOAuth, authHelper)
	userHandler := handlers.NewUserHandler(repo, pageHelper, texts.PasswordHint, texts.PasswordInfoCardTitle, texts.PasswordInfoCardItems)

	router := mux.NewRouter()
	router.Use(logMiddleware.LogRequests)
	router.Use(authMiddleware.EnsureAuth)

	router.HandleFunc("/", homeHandler.ServeHome).Methods("GET")
	router.HandleFunc("/auth", authHandler.RootHandler).Methods("GET")
	router.HandleFunc("/auth/google/callback", authHandler.GoogleCallbackHandler).Methods("GET")
	router.HandleFunc("/user", userHandler.List)
	router.HandleFunc("/user/new", userHandler.Enroll)
	router.HandleFunc("/user/{email}", userHandler.View)
	router.HandleFunc("/user/{email}/password", userHandler.Renew)
	router.HandleFunc("/user/{email}/delete", userHandler.Delete)
	router.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
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
