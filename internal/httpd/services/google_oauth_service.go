package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/p-l/fringe/internal/httpd/helpers"
)

type GoogleOAuthService struct {
	ClientID          string
	ClientSecret      string
	ClientCallbackURL string
	httpClient        *http.Client
}

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	HD            string `json:"hd"`
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

//

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

	log.Printf("Google Reply: %s", body)

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

func (g *GoogleOAuthService) AuthenticateUserWithToken(ctx context.Context, tokenType string, token string) (*GoogleUserInfo, error) {
	userInfo, err := g.fetchGoogleUserInfoWithToken(ctx, tokenType, token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGoogleAuthenticationFailed, err)
	}

	return userInfo, nil
}
