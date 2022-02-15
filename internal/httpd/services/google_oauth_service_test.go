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
		googleUser, err := service.AuthenticateUserWithToken(context.Background(), "bearer", "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Refuse authentication when user info call fails", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusTeapot,
				Body:       ioutil.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithToken(context.Background(), "bearer", "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Refuses authentication on empty user info", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(`{ "sub": "a_sub", "picture": "https://profile/picture/url", "hd": "domain.com" }`)),
				Header: make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithToken(context.Background(), "bearer", "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Refuses authentication on malformed user info", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(
					bytes.NewBufferString(`INVALID RESPONSE`)),
				Header: make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithToken(context.Background(), "bearer", "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Refuses authentication absent response", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
				Header:     make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithToken(context.Background(), "bearer", "code")
		assert.Error(t, err)
		assert.ErrorIs(t, err, services.ErrGoogleAuthenticationFailed)
		assert.Nil(t, googleUser)
	})

	t.Run("Accept authentication on user info", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{ "sub": "a_sub", "email": "email@domain.com", "email_verified": true, "picture": "https://profile/picture/url", "hd": "domain.com" }`)),
				Header: make(http.Header),
			}
		})

		service := services.NewGoogleOAuthService(client, "client_id", "client_secret", "https://redirect.url/somewhere/callback")
		googleUser, err := service.AuthenticateUserWithToken(context.Background(), "bearer", "code")
		assert.NoError(t, err)
		assert.NotNil(t, googleUser)
		assert.Equalf(t, "email@domain.com", googleUser.Email, "Expect userinfo email to match email in JSON")
	})
}
