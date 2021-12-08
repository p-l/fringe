package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/p-l/fringe/internal/http/helpers"
	"github.com/p-l/fringe/internal/repos"
)

type HomeHandler struct {
	UserRepo *repos.UserRepository
}

func NewHomeHandler(userRepo *repos.UserRepository) *HomeHandler {
	return &HomeHandler{
		UserRepo: userRepo,
	}
}

func (u *HomeHandler) ServeHome(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, success := helpers.AuthClaimsFromContext(httpRequest.Context())
	if !success {
		log.Printf("Home [%v]: Could not read context", httpRequest.RemoteAddr)
		http.Error(httpResponse, "could not retrieve user information", http.StatusInternalServerError)
	}

	log.Printf("Home [%v]: claims %v", httpRequest.RemoteAddr, claims.Email)
	// email := claims.Email

	user, err := u.UserRepo.UserWithEmail(claims.Email)
	if errors.Is(err, repos.ErrUserNotFound) {
		http.Redirect(httpResponse, httpRequest, "/user/new", http.StatusFound)

		return
	}

	if err != nil {
		log.Printf("Home [%v]: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "could not retrieve user information", http.StatusInternalServerError)
	}

	userPath := url.PathEscape("/user/" + user.Email)
	log.Printf("Home [%v]: Redirecting to user's (%s) page %s", httpRequest.RequestURI, user.Email, userPath)
	http.Redirect(httpResponse, httpRequest, userPath, http.StatusFound)
}
