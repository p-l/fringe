package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/p-l/fringe/internal/httpd/helpers"
)

type GoogleOAuthService struct {
	ClientID          string
	ClientSecret      string
	ClientCallbackURL string
	httpClient        *http.Client
}

type googleAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

var ErrGoogleAuthenticationFailed = errors.New("invalid response from google API")

func NewGoogleOAuthService(httpClient *http.Client, clientID string, clientSecret string, clientCallbackURL string) *GoogleOAuthService {
	return &GoogleOAuthService{
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		ClientCallbackURL: clientCallbackURL,
		httpClient:        httpClient,
	}
}

func (g *GoogleOAuthService) fetchGoogleTokenFromCallbackCode(ctx context.Context, code string) (auth *googleAuthResponse, err error) {
	postParams := url.Values{}
	postParams.Add("code", code)
	postParams.Add("client_id", g.ClientID)
	postParams.Add("client_secret", g.ClientSecret)
	postParams.Add("redirect_uri", g.ClientCallbackURL)
	postParams.Add("grant_type", "authorization_code")

	postBody := strings.NewReader(postParams.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", postBody)
	if err != nil {
		return nil, fmt.Errorf("fail to create http request to oauth2 api: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fail to post to oauth2 api: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 from https://oauth2.googleapis.com/token and got %d: %w", resp.StatusCode, ErrGoogleAuthenticationFailed)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read oauth2 response: %w", err)
	}

	var googleAuth googleAuthResponse
	if err = json.Unmarshal(body, &googleAuth); err != nil {
		return nil, fmt.Errorf("failed to parse oauth2 response: %w", err)
	}

	return &googleAuth, nil
}

func (g *GoogleOAuthService) fetchGoogleUserInfoWithToken(ctx context.Context, tokenType string, token string) (userInfo *GoogleUserInfo, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to request userinfo: %w", err)
	}

	req.Header.Add("Authorization", fmt.Sprintf("%s %s", tokenType, token))

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request userinfo: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected 200 from https://openidconnect.googleapis.com/v1/userinfo and got %d: %w", resp.StatusCode, ErrGoogleAuthenticationFailed)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo: %w", err)
	}

	var googleUserInfo GoogleUserInfo

	err = json.Unmarshal(body, &googleUserInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	if !helpers.IsEmailValid(googleUserInfo.Email) {
		return nil, fmt.Errorf("invalid email '%s': %w", googleUserInfo.Email, ErrGoogleAuthenticationFailed)
	}

	return &googleUserInfo, nil
}

func (g *GoogleOAuthService) AuthenticateUserWithCode(ctx context.Context, code string) (*GoogleUserInfo, error) {
	googleAuthResponse, err := g.fetchGoogleTokenFromCallbackCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleAuthenticationFailed, err)
	}

	userInfo, err := g.fetchGoogleUserInfoWithToken(ctx, googleAuthResponse.TokenType, googleAuthResponse.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleAuthenticationFailed, err)
	}

	return userInfo, nil
}

func (g *GoogleOAuthService) RedirectURL() string {
	// At the moment only support Google Authentication so redirect
	googleAuthScope := "https://www.googleapis.com/auth/userinfo.email"
	googleAuthURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s",
		url.QueryEscape(g.ClientID),
		url.QueryEscape(g.ClientCallbackURL),
		url.QueryEscape(googleAuthScope))

	return googleAuthURL
}
