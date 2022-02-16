package handlers_test

import (
	"bytes"
	"encoding/json"
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

func TestAuthHandler_Login(t *testing.T) {
	t.Parallel()

	t.Run("Refuse authentication to outside domains", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{ "sub": "a_sub", "email": "email@domain.com", "email_verified": true, "picture": "https://profile/picture/url", "hd": "domain.com" }`)),
				Header: make(http.Header),
			}
		})

		googleOAuth := services.NewGoogleOAuthService(client, "id", "secret", "callback")
		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		userRepo := mocks.NewMockUserRepository(t)

		authHandler := handlers.NewAuthHandler(userRepo, googleOAuth, authHelper)

		loginRequest := handlers.LoginRequest{
			AccessToken: "test_token",
			TokenType:   "token",
		}
		jsonBytes, err := json.Marshal(loginRequest)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/auth/", authHandler.Login)
		router.ServeHTTP(res, req)

		// Ensure Access Deny for domain reason
		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)

		body := res.Body.String()
		assert.Contains(t, body, "Domain is not allowed")
	})

	t.Run("Returns token on successful auth", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body: ioutil.NopCloser(bytes.NewBufferString(
					`{ "sub": "a_sub", "email": "email@test.com", "email_verified": true, "picture": "https://profile/picture/url", "name": "Person Name", "hd": "domain.com" }`)),
				Header: make(http.Header),
			}
		})

		googleOAuth := services.NewGoogleOAuthService(client, "id", "secret", "callback")
		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		userRepo := mocks.NewMockUserRepository(t)
		authHandler := handlers.NewAuthHandler(userRepo, googleOAuth, authHelper)

		loginRequest := handlers.LoginRequest{
			AccessToken: "test_token",
			TokenType:   "token_type",
		}
		jsonBytes, err := json.Marshal(loginRequest)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/auth/", authHandler.Login)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		// includes bearer token in response
		var response handlers.LoginResponse
		err = json.Unmarshal(res.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.NotEmpty(t, response.Token)
		assert.Equal(t, response.TokenType, "Bearer")
	})

	t.Run("Returns error on invalid post data", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       nil,
				Header:     make(http.Header),
			}
		})

		googleOAuth := services.NewGoogleOAuthService(client, "id", "secret", "callback")
		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		userRepo := mocks.NewMockUserRepository(t)
		authHandler := handlers.NewAuthHandler(userRepo, googleOAuth, authHelper)

		req := httptest.NewRequest(http.MethodPost, "/auth/", nil)
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/auth/", authHandler.Login)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	})

	t.Run("Returns error on refused Google user info", func(t *testing.T) {
		t.Parallel()

		client := mocks.NewMockHTTPClient(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       nil,
				Header:     make(http.Header),
			}
		})

		googleOAuth := services.NewGoogleOAuthService(client, "id", "secret", "callback")
		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		userRepo := mocks.NewMockUserRepository(t)
		authHandler := handlers.NewAuthHandler(userRepo, googleOAuth, authHelper)

		loginRequest := handlers.LoginRequest{
			AccessToken: "invalid_test_token",
			TokenType:   "token_type",
		}
		jsonBytes, err := json.Marshal(loginRequest)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/auth/", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")

		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/auth/", authHandler.Login)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	})
}
