package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/repos"
	"github.com/sethvargo/go-password/password"
)

type UserHandler struct {
	userRepo     *repos.UserRepository
	pageHelper   *helpers.PageHelper
	passwordHint string
	infoTitle    string
	infoItems    []string
}

type UserTemplateData struct {
	Email        string
	Password     string
	PasswordHint string
	InfoTitle    string
	InfoItems    []string
}

const (
	newUserPasswordLen          = 24
	newUserPasswordNumOfDigits  = 2
	newUserPasswordNumOfSymbols = 2
)

func NewUserHandler(userRepo *repos.UserRepository, pageHelper *helpers.PageHelper, paaswordHint string, infoTitle string, infoItems []string) *UserHandler {
	return &UserHandler{
		userRepo:     userRepo,
		pageHelper:   pageHelper,
		passwordHint: paaswordHint,
		infoTitle:    infoTitle,
		infoItems:    infoItems,
	}
}

func (u *UserHandler) List(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// claims, _ := httpRequest.Context().Value(helpers.AuthClaimsContextKey).(helpers.AuthClaims)
	// email = claims.Email

	http.Error(httpResponse, "Not implemented yet", http.StatusNotImplemented)
}

func (u *UserHandler) View(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	email := claims.Email
	data := UserTemplateData{
		Email:        email,
		PasswordHint: u.passwordHint,
		InfoTitle:    u.infoTitle,
		InfoItems:    u.infoItems,
	}

	err := u.pageHelper.RenderPage(httpResponse, "user/show.gohtml", data)
	if err != nil {
		log.Printf("User/View [%v]: template error: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Missing Template File", http.StatusInternalServerError)

		return
	}
}

func (u *UserHandler) Renew(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// Password will be return and should not be stored in cache
	forceHTTPNoCache(httpResponse)

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	email := claims.Email

	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	pwd, err := password.Generate(newUserPasswordLen, newUserPasswordNumOfDigits, newUserPasswordNumOfSymbols, false, false)
	if err != nil {
		log.Printf("User/New [%v]: Fail to renew password for %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	updated, err := u.userRepo.UpdateUserPassword(email, pwd)
	if err != nil {
		log.Printf("User/New [%v]: Fail to renew password for %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	if !updated {
		log.Printf("User/New [%v]: Fail to renew password for %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	data := UserTemplateData{
		Email:        email,
		Password:     pwd,
		PasswordHint: u.passwordHint,
		InfoTitle:    u.infoTitle,
		InfoItems:    u.infoItems,
	}

	err = u.pageHelper.RenderPage(httpResponse, "user/password.gohtml", data)
	if err != nil {
		log.Printf("User/View [%v]: template error: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Missing Template File", http.StatusInternalServerError)

		return
	}
}

func (u *UserHandler) Enroll(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// Password will be return and should not be stored in cache
	forceHTTPNoCache(httpResponse)

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

func (u *UserHandler) Delete(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// claims, _ := httpRequest.Context().Value(helpers.AuthClaimsContextKey).(helpers.AuthClaims)
	// email = claims.Email

	http.Error(httpResponse, "Not implemented yet", http.StatusNotImplemented)
}

func forceHTTPNoCache(httpResponse http.ResponseWriter) {
	httpResponse.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate")
	httpResponse.Header().Add("Pragma", "no-cache")
	httpResponse.Header().Add("Expires", "0")
}
