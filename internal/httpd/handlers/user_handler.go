package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mrz1836/go-sanitize"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/repos"
	"github.com/sethvargo/go-password/password"
)

type UserHandler struct {
	userRepo   *repos.UserRepository
	authHelper *helpers.AuthHelper
	pageHelper *helpers.PageHelper
}

type UserViewTemplateData struct {
	Email            string
	Password         string
	ShowUserListLink bool
}

type UserListTemplateData struct {
	Users     []repos.User
	Page      int
	LastPage  bool
	FirstPage bool
}

const (
	newUserPasswordLen          = 24
	newUserPasswordNumOfDigits  = 2
	newUserPasswordNumOfSymbols = 2
)

func NewUserHandler(userRepo *repos.UserRepository, authHelper *helpers.AuthHelper, pageHelper *helpers.PageHelper) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		authHelper: authHelper,
		pageHelper: pageHelper,
	}
}

func claimsAllowsForUserPage(claims *helpers.AuthClaims, targetEmail string) bool {
	return claims != nil && (strings.EqualFold(claims.Email, targetEmail) || claims.IsAdmin())
}

func forceHTTPNoCache(httpResponse http.ResponseWriter) {
	httpResponse.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate")
	httpResponse.Header().Add("Pragma", "no-cache")
	httpResponse.Header().Add("Expires", "0")
}

func (u *UserHandler) List(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()
	pageNumber := 0

	pageQueried := httpRequest.URL.Query().Get("page")

	if len(pageQueried) > 0 {
		pageNumber, err := strconv.Atoi(pageQueried)
		if err != nil {
			log.Printf("User/List [%v]: Could not parse page number '%d', defaulting to 0", httpRequest.RemoteAddr, pageNumber)
		}
	}

	claims, ok := helpers.AuthClaimsFromContext(httpRequest.Context())
	if !ok {
		log.Printf("User/List [%v]: failed to retrieve claims", httpRequest.RemoteAddr)
		http.Error(httpResponse, "could not extract claims from context", http.StatusInternalServerError)

		return
	}

	if !claims.IsAdmin() {
		log.Printf("User/List [%v]: %s is not allowed to list users", httpRequest.RemoteAddr, claims.Email)
		http.Error(httpResponse, "not authorized to list users", http.StatusUnauthorized)

		return
	}

	users, err := u.userRepo.AllUsers(0, pageNumber)
	if !errors.Is(err, repos.ErrUserNotFound) && err != nil {
		log.Printf("User/List [%v]: could not retrieve users: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "could not extract claims from context", http.StatusInternalServerError)

		return
	}

	data := UserListTemplateData{
		Users:     users,
		Page:      pageNumber,
		LastPage:  true,
		FirstPage: false,
		// LastPage:  len(users) < repos.UserRepositoryListMaxLimit,
		// FirstPage: pageNumber <= 1,
	}

	err = u.pageHelper.RenderPage(httpResponse, "user/list.gohtml", data)
	if err != nil {
		log.Printf("User/List [%v]: template error: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Missing Template File", http.StatusInternalServerError)

		return
	}
}

func (u *UserHandler) View(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if !claimsAllowsForUserPage(claims, email) {
		log.Printf("User/View [%v]: %s requested %s page without admin rights", httpRequest.RemoteAddr, claims.Email, email)
		http.Redirect(httpResponse, httpRequest, fmt.Sprintf("/user/%s/", claims.Email), http.StatusFound)

		return
	}

	data := UserViewTemplateData{
		Email:            email,
		ShowUserListLink: claims.IsAdmin(),
	}

	err := u.pageHelper.RenderPage(httpResponse, "user/show.gohtml", data)
	if err != nil {
		log.Printf("User/View [%v]: template error: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Missing Template File", http.StatusInternalServerError)
	}
}

func (u *UserHandler) Renew(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// Password will be return and should not be stored in cache
	forceHTTPNoCache(httpResponse)

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if !claimsAllowsForUserPage(claims, email) {
		log.Printf("User/Renew [%v]: %s cannot renew password for %s", httpRequest.RemoteAddr, claims.Email, email)
		http.Redirect(httpResponse, httpRequest, fmt.Sprintf("/user/%s/", claims.Email), http.StatusFound)

		return
	}

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

	u.renderPasswordPage(httpResponse, httpRequest, email, pwd)
}

func (u *UserHandler) renderPasswordPage(httpResponse http.ResponseWriter, httpRequest *http.Request, email string, password string) {
	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())

	data := UserViewTemplateData{
		Email:            email,
		Password:         password,
		ShowUserListLink: claims.IsAdmin(),
	}

	err := u.pageHelper.RenderPage(httpResponse, "user/password.gohtml", data)
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
	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if !claimsAllowsForUserPage(claims, email) {
		log.Printf("User/Renew [%v]: %s cannot enroll another user (%s)", httpRequest.RemoteAddr, claims.Email, email)
		http.Redirect(httpResponse, httpRequest, fmt.Sprintf("/user/%s/", claims.Email), http.StatusFound)

		return
	}

	if !helpers.IsEmailValid(email) || !u.authHelper.InAllowedDomain(email) {
		log.Printf("User/New [%v]: Invalid email: %s", httpRequest.RemoteAddr, email)
		http.Error(httpResponse, "failed to create user", http.StatusInternalServerError)

		return
	}

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

	if errors.Is(err, repos.ErrUserAlreadyExist) {
		log.Printf("User/New [%v]: user %s exists!", httpRequest.RemoteAddr, user.Email)
		http.Redirect(httpResponse, httpRequest, fmt.Sprintf("/user/%s/", email), http.StatusFound)

		return
	}

	log.Printf("User/New [%v]: user %s created!", httpRequest.RemoteAddr, user.Email)
	u.renderPasswordPage(httpResponse, httpRequest, email, pwd)
}

func (u *UserHandler) Delete(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if !claims.IsAdmin() {
		log.Printf("User/Delete [%v]: %s attempted to delete user %s without permission", httpRequest.RemoteAddr, claims.Email, email)
		http.Error(httpResponse, "not authorized to delete users", http.StatusUnauthorized)

		return
	}

	if !helpers.IsEmailValid(email) {
		log.Printf("User/Delete [%v]: Invalid email: %s", httpRequest.RemoteAddr, email)
		http.Redirect(httpResponse, httpRequest, "/", http.StatusFound)

		return
	}

	err := u.userRepo.DeleteUser(email)
	if err != nil && !errors.Is(err, repos.ErrUserNotFound) {
		log.Printf("User/Delete [%v]: failed to delete %s : %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "operation failed", http.StatusInternalServerError)

		return
	}

	if errors.Is(err, repos.ErrUserNotFound) {
		log.Printf("User/Delete [%v]: user %s doesn't exist", httpRequest.RemoteAddr, email)
	} else {
		log.Printf("User/Delete [%v]: user %s DELETED", httpRequest.RemoteAddr, email)
	}

	http.Redirect(httpResponse, httpRequest, "/user/", http.StatusFound)
}
