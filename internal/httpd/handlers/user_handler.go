package handlers

import (
	"encoding/json"
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
}

type UserCreateRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UserResponse struct {
	Email             string `json:"email"`
	Name              string `json:"name"`
	Picture           string `json:"picture"`
	Password          string `json:"password"`
	PasswordUpdatedAt int64  `json:"password_updated_at"`
	LastSeenAt        int64  `json:"last_seen_at"`
}

type UserActionResponse struct {
	Result string        `json:"result"`
	User   *UserResponse `json:"user"`
}

const (
	actionResultSuccess  = "success"
	actionResultFailed   = "failed"
	actionResultNotFound = "not_found"
	actionResultExists   = "exists"
)

const (
	newUserPasswordLen          = 24
	newUserPasswordNumOfDigits  = 2
	newUserPasswordNumOfSymbols = 2
)

func NewUserHandler(userRepo *repos.UserRepository, authHelper *helpers.AuthHelper) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		authHelper: authHelper,
	}
}

func claimsAllowsForUserPage(claims *helpers.AuthClaims, targetEmail string) bool {
	return claims != nil && (strings.EqualFold(claims.Email, targetEmail) || claims.IsAdmin())
}

func renderResponseJSONResponse(httpResponse http.ResponseWriter, httpRequest *http.Request, jsonResponse []byte, jsonErr error) {
	if jsonErr != nil {
		log.Printf("User [src:%v] failed to encode response: %v", httpRequest.RemoteAddr, jsonErr)
		http.Error(httpResponse, jsonErr.Error(), http.StatusInternalServerError)

		return
	}

	httpResponse.Header().Set("Content-Type", "application/json")
	httpResponse.Header().Add("Cache-Control", "no-store, no-cache, must-revalidate")
	httpResponse.Header().Add("Pragma", "no-cache")
	httpResponse.Header().Add("Expires", "0")

	_, err := httpResponse.Write(jsonResponse)
	if err != nil {
		log.Printf("User [src:%v] failed to send response: %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)
	}
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

	jsonResponse, jsonErr := json.Marshal(response)
	renderResponseJSONResponse(httpResponse, httpRequest, jsonResponse, jsonErr)
}

func renderActionResponse(httpResponse http.ResponseWriter, httpRequest *http.Request, response *UserActionResponse) {
	jsonResponse, jsonErr := json.Marshal(response)
	renderResponseJSONResponse(httpResponse, httpRequest, jsonResponse, jsonErr)
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

	jsonResponse, jsonErr := json.Marshal(returnedUsers)
	renderResponseJSONResponse(httpResponse, httpRequest, jsonResponse, jsonErr)
}

func isAuthorizedAdminRequest(httpRequest *http.Request) (authorized bool) {
	claims, ok := helpers.AuthClaimsFromContext(httpRequest.Context())
	if !ok {
		log.Printf("User [%v]: failed to retrieve claims", httpRequest.RemoteAddr)

		return false
	}

	if !claims.IsAdmin() {
		log.Printf("User [%v]: %s is not allowed to perform action", httpRequest.RemoteAddr, claims.Email)

		return false
	}

	return true
}

func createNewUser(repo *repos.UserRepository, email string, name string, picture string) (user *repos.User, userPassword *string, err error) {
	pwd, err := password.Generate(newUserPasswordLen, newUserPasswordNumOfDigits, newUserPasswordNumOfSymbols, false, false)
	if err != nil {
		return nil, nil, fmt.Errorf("password generation failed: %w", err)
	}

	user, err = repo.Create(email, name, picture, pwd)
	if err != nil {
		return nil, nil, fmt.Errorf("user creation failed: %w", err)
	}

	return user, &pwd, nil
}

func (u *UserHandler) List(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	var err error

	pageNumber := 0
	pageSize := 10
	pageQueried := sanitize.AlphaNumeric(httpRequest.URL.Query().Get("page"), false)
	perPage := sanitize.AlphaNumeric(httpRequest.URL.Query().Get("per_page"), false)
	searchQuery := sanitize.SingleLine(httpRequest.URL.Query().Get("search"))

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
			pageSize = 10
			log.Printf("User/List [%v]: Could not parse per_page '%s', defaulting to %d", httpRequest.RemoteAddr, perPage, pageSize)
		}
	}

	if !isAuthorizedAdminRequest(httpRequest) {
		http.Error(httpResponse, "not authorized to list users", http.StatusUnauthorized)

		return
	}

	var users []repos.User
	if len(searchQuery) > 0 {
		users, err = u.userRepo.FindAllMatching(searchQuery, pageSize, pageNumber)
	} else {
		users, err = u.userRepo.AllUsers(pageSize, pageNumber)
	}

	if err != nil {
		if errors.Is(err, repos.ErrUserNotFound) {
			log.Printf("User/List [%v]: query for users (search=%s) and had no results", httpRequest.RemoteAddr, searchQuery)
			users = []repos.User{}
		} else {
			log.Printf("User/List [%v]: could not get user list (page:%d, pageSize:%d search:%s): %v", httpRequest.RemoteAddr, pageNumber, pageSize, searchQuery, err)
			http.Error(httpResponse, "failed to query database", http.StatusInternalServerError)

			return
		}
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
	if errors.Is(err, repos.ErrUserNotFound) && strings.EqualFold(email, claims.Email) {
		newUser, pwd, err := createNewUser(u.userRepo, claims.Email, claims.Name, claims.Picture)
		if err != nil {
			log.Printf("User/View [%v]: %s requested %s but failed: %v", httpRequest.RemoteAddr, claims.Email, email, err)
			http.Error(httpResponse, err.Error(), http.StatusNotFound)

			return
		}

		userPassword = *pwd
		user = newUser
	} else if err != nil {
		log.Printf("User/View [%v]: %s requested %s but failed: %v", httpRequest.RemoteAddr, claims.Email, email, err)
		http.Error(httpResponse, err.Error(), http.StatusNotFound)

		return
	}

	renderUserResponse(httpResponse, httpRequest, user, userPassword)
}

func (u *UserHandler) Renew(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

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

	vars := mux.Vars(httpRequest)
	email := sanitize.Email(vars["email"], false)

	if !isAuthorizedAdminRequest(httpRequest) {
		http.Error(httpResponse, "not authorized to delete user", http.StatusUnauthorized)

		return
	}

	if !helpers.IsEmailValid(email) {
		log.Printf("User/Delete [%v]: Invalid email: %s", httpRequest.RemoteAddr, email)
		http.Error(httpResponse, "invalid email", http.StatusBadRequest)

		return
	}

	response := UserActionResponse{}

	err := u.userRepo.Delete(email)
	if err != nil {
		log.Printf("User/Delete [%v]: failed to delete: %s : %v", httpRequest.RemoteAddr, email, err)
		response.Result = actionResultFailed

		if errors.Is(err, repos.ErrUserNotFound) {
			response.Result = actionResultNotFound
		}
	} else {
		log.Printf("User/Delete [%v]: user %s Deleted", httpRequest.RemoteAddr, email)

		response.Result = actionResultSuccess
	}

	renderActionResponse(httpResponse, httpRequest, &response)
}

func (u *UserHandler) Create(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	var request UserCreateRequest
	decoder := json.NewDecoder(httpRequest.Body)

	err := decoder.Decode(&request)
	if err != nil {
		log.Printf("User/Create [src:%v] invalid post data %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Unable decode request", http.StatusBadRequest)

		return
	}

	email := sanitize.Email(request.Email, false)
	name := sanitize.SingleLine(sanitize.Punctuation(request.Name))

	if !isAuthorizedAdminRequest(httpRequest) {
		http.Error(httpResponse, "not authorized to create user", http.StatusUnauthorized)

		return
	}

	if !helpers.IsEmailValid(email) {
		log.Printf("User/Create [%v]: Invalid email: %s", httpRequest.RemoteAddr, email)
		http.Error(httpResponse, "invalid email", http.StatusBadRequest)

		return
	}

	if !helpers.IsEmailInDomain(email, u.authHelper.AllowedDomain) {
		log.Printf("User/Create [%v]: Invalid email: %s", httpRequest.RemoteAddr, email)
		http.Error(httpResponse, "email outside of authorized domain", http.StatusBadRequest)

		return
	}

	response := UserActionResponse{}

	user, pwd, err := createNewUser(u.userRepo, email, name, "")
	if err != nil {
		log.Printf("User/Create [%v]: failed to create: %s : %v", httpRequest.RemoteAddr, email, err)
		response.Result = actionResultExists
	} else {
		log.Printf("User/Create [%v]: user %s Created", httpRequest.RemoteAddr, email)

		response.Result = actionResultSuccess
		response.User = &UserResponse{
			Email:             user.Email,
			Name:              user.Name,
			Picture:           user.Picture,
			Password:          *pwd,
			PasswordUpdatedAt: user.PasswordUpdatedAt,
			LastSeenAt:        user.LastSeenAt,
		}
	}

	renderActionResponse(httpResponse, httpRequest, &response)
}
