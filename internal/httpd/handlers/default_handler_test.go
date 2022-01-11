package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/p-l/fringe/internal/mocks"
	"github.com/p-l/fringe/templates"
	"github.com/stretchr/testify/assert"
)

func TestDefaultHandler_Root(t *testing.T) {
	t.Parallel()

	t.Run("Does not process requests without claims", func(t *testing.T) {
		t.Parallel()

		pageHelper := helpers.NewPageHelper(templates.Files())
		userRepo := mocks.NewMockUserRepository(t)

		defaultHandler := handlers.NewDefaultHandler(userRepo, pageHelper)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/", defaultHandler.Root)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)
	})

	t.Run("Redirect if user is enrolled", func(t *testing.T) {
		t.Parallel()

		pageHelper := helpers.NewPageHelper(templates.Files())
		userRepo := mocks.NewMockUserRepository(t)

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware("/auth", []string{"/"}, []string{}, authHelper)

		defaultHandler := handlers.NewDefaultHandler(userRepo, pageHelper)

		users, err := userRepo.AllUsers(1, 0)
		if err != nil || len(users) < 1 {
			t.Fatalf("MockUserRepository doesn't contat users: %v", err)
		}

		user := users[0]
		claims := helpers.NewAuthClaims(user.Email, "")
		tokenCookie := authHelper.NewJWTCookieFromClaims(claims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		req.AddCookie(tokenCookie)

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)

		router.HandleFunc("/", defaultHandler.Root)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusFound, res.Result().StatusCode)

		location, _ := res.Result().Location()
		assert.Contains(t, location.Path, "/user/")
	})

	t.Run("Redirect if user is NOT enrolled", func(t *testing.T) {
		t.Parallel()

		pageHelper := helpers.NewPageHelper(templates.Files())
		userRepo := mocks.NewMockUserRepository(t)

		authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
		authMiddleware := middlewares.NewAuthMiddleware("/auth", []string{"/"}, []string{}, authHelper)
		defaultHandler := handlers.NewDefaultHandler(userRepo, pageHelper)
		claims := helpers.NewAuthClaims("user_does_not_exist@not_a_user.com", "")
		tokenCookie := authHelper.NewJWTCookieFromClaims(claims)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		req.AddCookie(tokenCookie)

		router := mux.NewRouter()
		router.Use(authMiddleware.EnsureAuth)
		router.HandleFunc("/", defaultHandler.Root)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusFound, res.Result().StatusCode)

		location, _ := res.Result().Location()
		assert.Contains(t, location.Path, "enroll")
	})
}
