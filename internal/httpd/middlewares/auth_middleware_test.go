package middlewares_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jaswdr/faker"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/stretchr/testify/assert"
)

const authPath = "/auth/"

func TestAuthMiddleware_EnsureAuth(t *testing.T) {
	t.Parallel()

	t.Run("Refuse Access to AuthPath if no token is present", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{}, authHelper)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			t.Errorf("Handler must not be called when middleware is called")
		})
		router.HandleFunc(authPath, func(writer http.ResponseWriter, request *http.Request) {})
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusForbidden, res.Result().StatusCode)
	})

	t.Run("skips token validation path in excludedPath", func(t *testing.T) {
		t.Parallel()

		testSkipAuthRootPath := "/skip-auth"
		testSkipAuthSubPath := testSkipAuthRootPath + "/test"
		protectedPaths := []string{"/"}
		skipPaths := []string{testSkipAuthRootPath}
		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(testSkipAuthRootPath, protectedPaths, skipPaths, authHelper)

		req := httptest.NewRequest(http.MethodGet, testSkipAuthSubPath, nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			t.Errorf("Root Handler must not be called")
		})
		router.HandleFunc(testSkipAuthSubPath, func(writer http.ResponseWriter, request *http.Request) {
			assert.Equal(t, testSkipAuthSubPath, request.RequestURI)
			writer.WriteHeader(http.StatusTeapot)
		})
		router.ServeHTTP(res, req)

		// Make sure the handler ran by testing for a specific result code.
		assert.Equal(t, http.StatusTeapot, res.Result().StatusCode)
	})

	t.Run("let valid tokens through", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware("/auth/", []string{"/"}, []string{"/no-auth"}, authHelper)

		validClaims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		// Force expiry to be 1 minute in the future
		validClaims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Minute).Unix()
		validToken := authHelper.NewJWTSignedString(validClaims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusTeapot)
		})
		router.ServeHTTP(res, req)

		// Make sure the handler ran by testing for a specific result code.
		assert.Equal(t, http.StatusTeapot, res.Result().StatusCode)
	})

	t.Run("rejects expired tokens", func(t *testing.T) {
		t.Parallel()

		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{"/no-auth"}, authHelper)

		validClaims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		// Force expiry to be 1 minute ago
		validClaims.StandardClaims.ExpiresAt = time.Now().Add(-1 * time.Minute).Unix()
		validToken := authHelper.NewJWTSignedString(validClaims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", validToken))
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			t.Errorf("Root Handler must not be called")
		})
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusForbidden, res.Result().StatusCode)
	})

	t.Run("rejects invalid token type strings", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{"/no-auth"}, authHelper)

		validClaims := helpers.NewAuthClaims(fake.Internet().Email(), fake.Person().Name(), fake.Internet().URL(), "")
		validClaims.StandardClaims.ExpiresAt = time.Now().Add(5 * time.Minute).Unix()
		validToken := authHelper.NewJWTSignedString(validClaims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearing %s", validToken))
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			t.Errorf("Root Handler must not be called")
		})
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusForbidden, res.Result().StatusCode)
	})

	t.Run("rejects missing token in authorization strings", func(t *testing.T) {
		t.Parallel()

		authPath := "/auth/"
		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{"/no-auth"}, authHelper)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Add("Authorization", "Bearer ")
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			t.Errorf("Root Handler must not be called")
		})
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusForbidden, res.Result().StatusCode)
	})
}

func TestAuthMiddleware_IsProtected(t *testing.T) {
	t.Parallel()

	t.Run("returns false when path is in the excluded list", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{"/no-auth"}, authHelper)

		assert.False(t, authMiddleware.IsProtected("/no-auth"))
	})

	t.Run("returns true when path is in the excluded list", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{"/no-auth"}, authHelper)

		assert.True(t, authMiddleware.IsProtected("/"))
	})

	t.Run("returns true when path is inside the included list", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/require-auth"}, []string{"/no-auth"}, authHelper)

		assert.True(t, authMiddleware.IsProtected("/require-auth"))
		assert.True(t, authMiddleware.IsProtected("/require-auth/"))
		assert.True(t, authMiddleware.IsProtected("/require-auth/subpath/"))
	})

	t.Run("returns false when its not inside the included list and not part of excluded list", func(t *testing.T) {
		t.Parallel()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/require-auth"}, []string{"/no-auth"}, authHelper)

		assert.False(t, authMiddleware.IsProtected("/other-path"))
	})
}
