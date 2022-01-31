package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/services"
	"github.com/p-l/fringe/internal/repos"
)

type AuthHandler struct {
	authHelper  *helpers.AuthHelper
	googleOAuth *services.GoogleOAuthService
	userRepo    *repos.UserRepository
}

type LoginRequest struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type LoginResponse struct {
	TokenType string `json:"token_type"`
	Token     string `json:"token"`
	Duration  int64  `json:"duration"`
}

func NewAuthHandler(userRepo *repos.UserRepository, googleOAuthService *services.GoogleOAuthService, authHelper *helpers.AuthHelper) *AuthHandler {
	return &AuthHandler{
		authHelper:  authHelper,
		googleOAuth: googleOAuthService,
		userRepo:    userRepo,
	}
}

// Login validates Google token and create JWT if its valid.
func (a *AuthHandler) Login(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	var data LoginRequest
	decoder := json.NewDecoder(httpRequest.Body)

	err := decoder.Decode(&data)
	if err != nil {
		log.Printf("Auth [src:%v] invalid post data %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Unable to validate code", http.StatusUnauthorized)

		return
	}

	googleUserInfo, err := a.googleOAuth.AuthenticateUserWithToken(httpRequest.Context(), data.TokenType, data.AccessToken)
	if err != nil {
		log.Printf("Auth [src:%v] invalid token %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Unable to validate code", http.StatusUnauthorized)

		return
	}

	if !a.authHelper.InAllowedDomain(googleUserInfo.Email) {
		log.Printf("Auth [src:%v] email (%s) is not in allowed domain (%s)", httpRequest.RemoteAddr, googleUserInfo.Email, a.authHelper.AllowedDomain)
		http.Error(httpResponse, "Domain is not allowed", http.StatusUnauthorized)

		return
	}

	permissions := a.authHelper.PermissionsForEmail(googleUserInfo.Email)
	claims := helpers.NewAuthClaims(googleUserInfo.Email, googleUserInfo.Name, googleUserInfo.Picture, permissions)
	signedTokenString := a.authHelper.NewJWTSignedString(claims)
	duration := time.Unix(claims.ExpiresAt, 0).Unix() - time.Now().Unix()

	response := LoginResponse{TokenType: "Bearer", Token: signedTokenString, Duration: duration}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Auth [src:%v] failed to encode token (for %s): %v", httpRequest.RemoteAddr, claims.Email, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)

		return
	}

	httpResponse.Header().Set("Content-Type", "application/json")

	_, err = httpResponse.Write(jsonResponse)
	if err != nil {
		log.Printf("Auth [src:%v] failed to send response token for %s: %v", httpRequest.RemoteAddr, claims.Email, err)
		http.Error(httpResponse, err.Error(), http.StatusInternalServerError)

		return
	}

	// try to update the profile if the user exists
	_, _ = a.userRepo.UpdateProfile(claims.Email, claims.Name, claims.Picture)
}
