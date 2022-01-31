package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mrz1836/go-sanitize"
	"github.com/sethvargo/go-password/password"

	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/repos"
)

type UserHandler struct {
	userRepo   *repos.UserRepository
	authHelper *helpers.AuthHelper
	pageHelper *helpers.PageHelper
}

type UserResponse struct {
	Email             string `json:"email"`
	Name              string `json:"name"`
	Picture           string `json:"picture"`
	Password          string `json:"password"`
	PasswordUpdatedAt int64  `json:"password_updated_at"`
	LastSeenAt        int64  `json:"last_seen_at"`
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

	var err error

	pageNumber := 0
	pageSize := 10
	pageQueried := httpRequest.URL.Query().Get("page")
	perPage := httpRequest.URL.Query().Get("per_page")

	if len(pageQueried) > 0 {
		pageNumber, err = strconv.Atoi(pageQueried)
		if err != nil {
			pageNumber = 0
			log.Printf("User/List [%v]: Could not parse page '%s', defaulting to %d", httpRequest.RemoteAddr, pageQueried, pageNumber)
		}
	}

	if len(perPage) > 0 {
		pageSize, err = strconv.Atoi(perPage)
		if err != nil {
			pageSize := 10
			log.Printf("User/List [%v]: Could not parse per_page '%s', defaulting to %d", httpRequest.RemoteAddr, perPage, pageSize)
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

	users, err := u.userRepo.AllUsers(pageSize, pageNumber)
	if err != nil {
		log.Printf("User/List [%v]: %s could not get user list (page:%d, pageSize:%d)", httpRequest.RemoteAddr, claims.Email, pageNumber, pageSize)
		http.Error(httpResponse, "not authorized to list users", http.StatusUnauthorized)

		return
	}

	renderUserListResponse(httpResponse, httpRequest, users)
}

func (u *UserHandler) View(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if strings.EqualFold(email, "me") {
		email = claims.Email
	}

	if !claimsAllowsForUserPage(claims, email) {
		log.Printf("User/View [%v]: %s requested %s page without admin rights", httpRequest.RemoteAddr, claims.Email, email)
		http.Error(httpResponse, "Not allowed", http.StatusForbidden)

		return
	}

	userPassword := ""

	user, err := u.userRepo.FindByEmail(email)
	if errors.Is(err, repos.ErrUserNotFound) {
		userPassword, err = password.Generate(newUserPasswordLen, newUserPasswordNumOfDigits, newUserPasswordNumOfSymbols, false, false)
		if err != nil {
			log.Printf("User/New [%v]: Fail to create user %s: %v", httpRequest.RemoteAddr, email, err)
			http.Error(httpResponse, "failed to create user", http.StatusInternalServerError)

			return
		}

		user, err = u.userRepo.Create(claims.Email, claims.Name, claims.Picture, userPassword)
		if err != nil && !errors.Is(err, repos.ErrUserAlreadyExist) {
			log.Printf("User/New [%v]: Fail to create user %s: %v", httpRequest.RemoteAddr, email, err)
			http.Error(httpResponse, "failed to create user", http.StatusInternalServerError)

			return
		}
	} else if err != nil {
		log.Printf("User/View [%v]: %s requested %s but failed: %v", httpRequest.RemoteAddr, claims.Email, email, err)
		http.Error(httpResponse, err.Error(), http.StatusNotFound)

		return
	}

	renderUserResponse(httpResponse, httpRequest, user, userPassword)
}

func (u *UserHandler) Renew(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// Password will be return and should not be stored in cache
	forceHTTPNoCache(httpResponse)

	claims, _ := helpers.AuthClaimsFromContext(httpRequest.Context())
	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if strings.EqualFold(email, "me") {
		email = claims.Email
	}

	if !claimsAllowsForUserPage(claims, email) {
		log.Printf("User/Renew [%v]: %s cannot renew password for %s", httpRequest.RemoteAddr, claims.Email, email)
		http.Error(httpResponse, "Not allowed", http.StatusForbidden)

		return
	}

	// Generate a password that is 64 characters long with 10 digits, 10 symbols,
	// allowing upper and lower case letters, disallowing repeat characters.
	pwd, err := password.Generate(newUserPasswordLen, newUserPasswordNumOfDigits, newUserPasswordNumOfSymbols, false, false)
	if err != nil {
		log.Printf("User/Renew [%v]: Fail to renew password for %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	updated, err := u.userRepo.UpdatePassword(email, pwd)
	if err != nil {
		log.Printf("User/Renew [%v]: Fail to renew password for %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	if !updated {
		log.Printf("User/Renew [%v]: Fail to renew password for %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	user, err := u.userRepo.FindByEmail(email)
	if err != nil {
		log.Printf("User/Renew [%v]: Fail to get user after password renew %s: %v", httpRequest.RemoteAddr, email, err)
		http.Error(httpResponse, "failed to renew password", http.StatusInternalServerError)

		return
	}

	renderUserResponse(httpResponse, httpRequest, user, pwd)
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

	err := u.userRepo.Delete(email)
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

func renderUserResponse(httpResponse http.ResponseWriter, httpRequest *http.Request, user *repos.User, pwd string) {
	response := UserResponse{
		Email:             user.Email,
		Name:              user.Name,
		Picture:           user.Picture,
		LastSeenAt:        user.LastSeenAt,
		PasswordUpdatedAt: user.PasswordUpdatedAt,
		Password:          pwd,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("User [src:%v] failed to encode %s user response: %v", httpRequest.RemoteAddr, user.Email, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)

		return
	}

	httpResponse.Header().Set("Content-Type", "application/json")

	_, err = httpResponse.Write(jsonResponse)
	if err != nil {
		log.Printf("Auth [src:%v] failed to send response %s user response: %v", httpRequest.RemoteAddr, user.Email, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)
	}
}

func renderUserListResponse(httpResponse http.ResponseWriter, httpRequest *http.Request, users []repos.User) {
	returnedUsers := make([]UserResponse, 0, len(users))
	for _, user := range users {
		returnedUsers = append(returnedUsers, UserResponse{
			Email:             user.Email,
			Name:              user.Name,
			Picture:           user.Picture,
			PasswordUpdatedAt: user.PasswordUpdatedAt,
			LastSeenAt:        user.LastSeenAt,
			Password:          "",
		})
	}

	jsonResponse, err := json.Marshal(returnedUsers)
	if err != nil {
		log.Printf("User [src:%v] failed to encode user list: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)

		return
	}

	httpResponse.Header().Set("Content-Type", "application/json")

	_, err = httpResponse.Write(jsonResponse)
	if err != nil {
		log.Printf("Auth [src:%v] failed to send user list: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)
	}
}
