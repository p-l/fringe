package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/p-l/fringe/internal/http/helpers"
	"github.com/p-l/fringe/internal/repos"
	"github.com/sethvargo/go-password/password"
)

type UserHandler struct {
	userRepo *repos.UserRepository
}

func NewUserHandler(userRepo *repos.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

func (u *UserHandler) ServerIndex(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// claims, _ := httpRequest.Context().Value(helpers.AuthClaimsContextKey).(helpers.AuthClaims)
	// email = claims.Email

	http.Error(httpResponse, "Not implemented yet", http.StatusNotImplemented)
}

func (u *UserHandler) ServeView(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// claims, _ := httpRequest.Context().Value(helpers.AuthClaimsContextKey).(helpers.AuthClaims)
	// email = claims.Email

	http.Error(httpResponse, "Not implemented yet", http.StatusNotImplemented)
}

const (
	newUserPasswordLen          = 24
	newUserPasswordNumOfDigits  = 2
	newUserPasswordNumOfSymbols = 2
)

func (u *UserHandler) ServeNew(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	email := claims.Email

	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	pwd, err := password.Generate(newUserPasswordLen, newUserPasswordNumOfDigits, newUserPasswordNumOfSymbols, false, false)
	if err != nil {
		log.Printf("User/New [%v]: Fail to create user %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to create user", http.StatusInternalServerError)

		return
	}

	user, err := u.userRepo.CreateUser(email, pwd)
	if err != nil && !errors.Is(err, repos.ErrUserAlreadyExist) {
		log.Printf("User/New [%v]: Fail to create user %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to create user", http.StatusInternalServerError)

		return
	}

	log.Printf("User/New [%v]: created user %s", httpRequest.RemoteAddr, user.Email)
	userPath := url.PathEscape("/user/" + user.Email)
	http.Redirect(httpResponse, httpRequest, userPath, http.StatusFound)
}

func (u *UserHandler) ServeDelete(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// claims, _ := httpRequest.Context().Value(helpers.AuthClaimsContextKey).(helpers.AuthClaims)
	// email = claims.Email

	http.Error(httpResponse, "Not implemented yet", http.StatusNotImplemented)
}
