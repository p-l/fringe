package handlers

import (
	"log"
	"net/http"

	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/services"
)

type AuthHandler struct {
	authHelper  *helpers.AuthHelper
	googleOAuth *services.GoogleOAuthService
}

func NewAuthHandler(googleOAuthService *services.GoogleOAuthService, authHelper *helpers.AuthHelper) *AuthHandler {
	return &AuthHandler{
		authHelper:  authHelper,
		googleOAuth: googleOAuthService,
	}
}

// GoogleCallbackHandler handles "/auth/google/callback"
// the path portion of the GoogleOAuthClientConfig.RedirectURI.
func (a *AuthHandler) GoogleCallbackHandler(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	parsedQuery := httpRequest.URL.Query()
	code := parsedQuery.Get("code")

	googleUserInfo, err := a.googleOAuth.AuthenticateUserWithCode(httpRequest.Context(), code)
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
	claims := helpers.NewAuthClaims(googleUserInfo.Email, permissions)
	cookie := a.authHelper.NewJWTCookieFromClaims(claims)
	http.SetCookie(httpResponse, cookie)
	http.Redirect(httpResponse, httpRequest, "/", http.StatusFound)
}

// RootHandler handles "/auth"
// Redirect to Google Authentication page for the moment.
// NOTE: When more than one OAuth provider is supported the behaviour will change.
func (a *AuthHandler) RootHandler(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	googleAuthURL := a.googleOAuth.RedirectURL()

	http.Redirect(httpResponse, httpRequest, googleAuthURL, http.StatusFound)
}
