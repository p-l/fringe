package handlers_test

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/services"
	"github.com/p-l/fringe/internal/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_RootHandler(t *testing.T) {
	t.Parallel()

	t.Run("Redirects to google", func(t *testing.T) {
		t.Parallel()

		authPath := "/auth/"

		googleOAuth := services.NewGoogleOAuthService(nil, "id", "secret", "callback")
		authHandler := handlers.NewAuthHandler(googleOAuth, nil)

		req := httptest.NewRequest(http.MethodGet, authPath, nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc(authPath, authHandler.RootHandler)
		router.ServeHTTP(res, req)

		// Ensure redirect
		assert.Equal(t, http.StatusFound, res.Result().StatusCode)

		// URL domain is google.com
		location, _ := res.Result().Location()
		assert.Contains(t, location.Host, "google.com")
	})
}

const googleTokenAPIURL = "https://oauth2.googleapis.com/token"

func TestAuthHandler_GoogleCallbackHandler(t *testing.T) {
	t.Parallel()

	t.Run("Refuse authentication to outside domains", func(t *testing.T) {
		t.Parallel()

		callbackPath := "/auth/google/callback"

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

		googleOAuth := services.NewGoogleOAuthService(client, "id", "secret", "callback")
		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})

		authHandler := handlers.NewAuthHandler(googleOAuth, authHelper)

		req := httptest.NewRequest(http.MethodGet, callbackPath, nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc(callbackPath, authHandler.GoogleCallbackHandler)
		router.ServeHTTP(res, req)

		// Ensure Access Deny for domain reason
		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)

		body := res.Body.String()
		assert.Contains(t, body, "Domain is not allowed")
	})

	t.Run("Set cookie on successful auth", func(t *testing.T) {
		t.Parallel()

		callbackPath := "/auth/"

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
						`{ "sub": "a_sub", "email": "email@test.com", "email_verified": true, "picture": "https://profile/picture/url", "hd": "domain.com" }`)),
					Header: make(http.Header),
				}
			}
		})

		googleOAuth := services.NewGoogleOAuthService(client, "id", "secret", "callback")
		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})

		authHandler := handlers.NewAuthHandler(googleOAuth, authHelper)

		req := httptest.NewRequest(http.MethodGet, callbackPath, nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc(callbackPath, authHandler.GoogleCallbackHandler)
		router.ServeHTTP(res, req)

		// Redirect to protected resource
		assert.Equal(t, http.StatusFound, res.Result().StatusCode)

		// Included the cookie
		cookies := res.Result().Cookies()
		assert.NotNil(t, cookies)
		assert.Equal(t, 1, len(cookies))

		authCookie := cookies[0]
		assert.Equal(t, "token", authCookie.Name)
	})
}
