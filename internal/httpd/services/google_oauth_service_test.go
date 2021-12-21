package services_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/p-l/fringe/internal/httpd/services"
	"github.com/p-l/fringe/internal/mocks"
	"github.com/stretchr/testify/assert"
)

const googleTokenAPIURL = "https://oauth2.googleapis.com/token"

func TestGoogleOAuthService_AuthenticateUserWithCode(t *testing.T) {
	t.Parallel()

	t.Run("Refuse authentication when token call fails", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusTeapot,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithCode(context.Background(), "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Refuse authentication when user info call fails", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			url := req.URL.String()
			switch url {
			case googleTokenAPIURL:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(`{ "access_token": "an_access_token", "expires_in": 42, "scope": "a_scope", "token_type": "bearer_test", "id_token": "an_id_token" }`)),
					Header:     make(http.Header),
				}

			default:
				return &http.Response{
					StatusCode: http.StatusTeapot,
					Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
					Header:     make(http.Header),
				}
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithCode(context.Background(), "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Refuses authentication on malformed user info", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			url := req.URL.String()
			switch url {
			case googleTokenAPIURL:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(`{ "access_token": "an_access_token", "expires_in": 42, "scope": "a_scope", "token_type": "bearer_test", "id_token": "an_id_token" }`)),
					Header:     make(http.Header),
				}
			default:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(
						bytes.NewBufferString(`{ "sub": "a_sub", "picture": "https://profile/picture/url", "hd": "domain.com" }`)),
					Header: make(http.Header),
				}
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithCode(context.Background(), "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Accept authentication on user info", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			url := req.URL.String()
			switch url {
			case googleTokenAPIURL:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(`{ "access_token": "an_access_token", "expires_in": 42, "scope": "a_scope", "token_type": "bearer_test", "id_token": "an_id_token" }`)),
					Header:     make(http.Header),
				}
			default:
				return &http.Response{
					StatusCode: http.StatusOK,
					Body: ioutil.NopCloser(bytes.NewBufferString(
						`{ "sub": "a_sub", "email": "email@domain.com", "email_verified": true, "picture": "https://profile/picture/url", "hd": "domain.com" }`)),
					Header: make(http.Header),
				}
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithCode(context.Background(), "code")
		assert.NoError(t, err)
		assert.NotNil(t, googleUser)
		assert.Equalf(t, "email@domain.com", googleUser.Email, "Expect userinfo email to match email in JSON")
	})
}

func TestGoogleOAuthService_RedirectURL(t *testing.T) {
	t.Parallel()
	t.Run("RedirectURL Includes userinfo.email scope at minimum", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusTeapot,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		url := service.RedirectURL()
		assert.Contains(t, url, "userinfo.email")
	})
}
