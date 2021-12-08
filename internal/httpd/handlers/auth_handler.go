package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/p-l/fringe/internal/httpd/helpers"
)

type GoogleOAuthClientConfig struct {
	ID          string
	Secret      string
	RedirectURL string
}

type AuthHandler struct {
	authHelper         *helpers.AuthHelper
	googleClientConfig GoogleOAuthClientConfig
	allowedDomain      string
}

type googleAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
}

type googleUserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

func NewAuthHandler(allowedDomain string, googleConfig GoogleOAuthClientConfig, authHelper *helpers.AuthHelper) *AuthHandler {
	return &AuthHandler{
		authHelper:         authHelper,
		googleClientConfig: googleConfig,
		allowedDomain:      allowedDomain,
	}
}

func fetchGoogleTokenFromCallbackCode(code string, googleConfig GoogleOAuthClientConfig) (auth googleAuthResponse, err error) {
	postParams := url.Values{}
	postParams.Add("code", code)
	postParams.Add("client_id", googleConfig.ID)
	postParams.Add("client_secret", googleConfig.Secret)
	postParams.Add("redirect_uri", googleConfig.RedirectURL)
	postParams.Add("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", postParams)
	if err != nil {
		return googleAuthResponse{}, fmt.Errorf("fail to post to oauth2 api: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return googleAuthResponse{}, fmt.Errorf("failed to read oauth2 response: %w", err)
	}

	var googleAuth googleAuthResponse
	if err = json.Unmarshal(body, &googleAuth); err != nil {
		return googleAuthResponse{}, fmt.Errorf("failed to parse oauth2 response: %w", err)
	}

	return googleAuth, nil
}

func fetchGoogleUserInfoWithToken(ctx context.Context, tokenType string, token string) (userInfo googleUserInfoResponse, err error) {
	var googleUserInfo googleUserInfoResponse

	req, err := http.NewRequestWithContext(ctx, "GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("failed to request userinfo: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", tokenType, token))

	userInfoResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("failed to request userinfo: %w", err)
	}

	defer func() { _ = userInfoResponse.Body.Close() }()

	userInfoBody, err := ioutil.ReadAll(userInfoResponse.Body)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("failed to read userinfo: %w", err)
	}

	err = json.Unmarshal(userInfoBody, &googleUserInfo)
	if err != nil {
		return googleUserInfoResponse{}, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	return googleUserInfo, nil
}

// GoogleCallbackHandler handles "/auth/google/callback"
// the path portion of the GoogleOAuthClientConfig.RedirectURI.
func (a *AuthHandler) GoogleCallbackHandler(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	parsedQuery := httpRequest.URL.Query()
	code := parsedQuery.Get("code")

	googleAuth, err := fetchGoogleTokenFromCallbackCode(code, a.googleClientConfig)
	if err != nil {
		log.Printf("Auth [src:%v] invalid token %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Invalid code", http.StatusUnauthorized)

		return
	}

	googleUserInfo, err := fetchGoogleUserInfoWithToken(httpRequest.Context(), googleAuth.TokenType, googleAuth.AccessToken)
	if err != nil {
		log.Printf("Auth [src:%v] invalid token %v", httpRequest.RemoteAddr, err)
		http.Error(httpResponse, "Unable to confirm user info", http.StatusUnauthorized)

		return
	}

	if !strings.Contains(googleUserInfo.Email, "@"+a.allowedDomain) {
		log.Printf("Auth [src:%v] email (%s) is not in allowed domain (%s)", httpRequest.RemoteAddr, googleUserInfo.Email, a.allowedDomain)
		http.Error(httpResponse, fmt.Sprintf("User not in allowed domain %s", a.allowedDomain), http.StatusUnauthorized)

		return
	}

	claims := helpers.NewAuthClaims(googleUserInfo.Email)
	cookie := a.authHelper.NewJWTCookieFromClaims(claims)
	http.SetCookie(httpResponse, cookie)
	http.Redirect(httpResponse, httpRequest, "/", http.StatusFound)
}

// RootHandler handles "/auth"
// Redirect to Google Authentication page for the moment.
// NOTE: When more than one OAuth provider is supported the behaviour will change.
func (a *AuthHandler) RootHandler(httpResponse http.ResponseWriter, httpRequest *http.Request) {
	defer func() { _ = httpRequest.Body.Close() }()

	// At the moment only support Google Authentication so redirect
	googleAuthScope := "https://www.googleapis.com/auth/userinfo.email"
	googleAuthURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		url.QueryEscape(a.googleClientConfig.ID),
		url.QueryEscape(a.googleClientConfig.RedirectURL),
		url.QueryEscape(googleAuthScope))

	http.Redirect(httpResponse, httpRequest, googleAuthURL, http.StatusFound)
}
