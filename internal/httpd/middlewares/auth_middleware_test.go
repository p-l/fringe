package middlewares_test

import (
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

func TestEnsureAuth(t *testing.T) {
	t.Parallel()

	t.Run("redirect to AuthPath if no token is present", func(t *testing.T) {
		t.Parallel()

		authPath := "/test/auth"
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

		assert.Equal(t, http.StatusFound, res.Result().StatusCode)

		target, err := res.Result().Location()
		assert.Nil(t, err)
		assert.Equal(t, authPath, target.Path)
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

	t.Run("let valid and refreshed tokens through", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware("/auth/", []string{"/"}, []string{"/no-auth"}, authHelper)

		validClaims := helpers.NewAuthClaims(fake.Internet().Email(), "")
		// Force expiry to be 1 minute in the future
		validClaims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Minute).Unix()
		validTokenCookie := authHelper.NewJWTCookieFromClaims(validClaims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(validTokenCookie)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusTeapot)
		})
		router.ServeHTTP(res, req)

		// Make sure the handler ran by testing for a specific result code.
		assert.Equal(t, http.StatusTeapot, res.Result().StatusCode)

		// Ensure the token refreshed
		cookies := res.Result().Cookies()
		assert.NotNil(t, cookies)
		assert.Equal(t, 1, len(cookies))

		authCookie := cookies[0]
		assert.Equal(t, "token", authCookie.Name)
		assert.Greater(t, authCookie.Expires.Unix(), validClaims.StandardClaims.ExpiresAt)

		cookieClaims, err := authHelper.AuthClaimsFromSignedToken(authCookie.Value)
		assert.Nil(t, err)
		assert.Greater(t, cookieClaims.StandardClaims.ExpiresAt, validClaims.StandardClaims.ExpiresAt)
	})

	t.Run("rejects expired tokens", func(t *testing.T) {
		t.Parallel()
		fake := faker.New()

		authPath := "/auth/"
		authHelper := helpers.NewAuthHelper("@test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware(authPath, []string{"/"}, []string{"/no-auth"}, authHelper)

		validClaims := helpers.NewAuthClaims(fake.Internet().Email(), "")
		// Force expiry to be 1 minute ago
		validClaims.StandardClaims.ExpiresAt = time.Now().Add(-1 * time.Minute).Unix()
		validTokenCookie := authHelper.NewJWTCookieFromClaims(validClaims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(validTokenCookie)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			t.Errorf("Root Handler must not be called")
		})
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusFound, res.Result().StatusCode)

		// Ensure redirected to Auth path
		target, err := res.Result().Location()
		assert.Nil(t, err)
		assert.Equal(t, authPath, target.Path)

		// Ensure the token is removed
		cookies := res.Result().Cookies()
		assert.NotNil(t, cookies)
		assert.Equal(t, 1, len(cookies))

		authCookie := cookies[0]
		expectedCookie := authHelper.RemoveJWTCookie()
		assert.Equal(t, expectedCookie.Name, authCookie.Name)
		assert.Equal(t, time.Unix(0, 0).Unix(), authCookie.Expires.Unix())
	})
}
