package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/p-l/fringe/internal/mocks"
	"github.com/p-l/fringe/internal/repos"
	"github.com/p-l/fringe/templates"
	"github.com/stretchr/testify/assert"
)

const (
	adminEmail       = "admin@test.com"
	regularUserEmail = "user@test.com"
)

func createUserHandler(t *testing.T) (*handlers.UserHandler, *repos.UserRepository) {
	t.Helper()

	userRepo := mocks.NewMockUserRepository(t)

	for _, email := range []string{adminEmail, regularUserEmail} {
		_, err := userRepo.CreateUser(email, "test_pass")
		if err != nil {
			t.Fatalf("Could not add admin user to test database: %v", err)
		}
	}

	authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})
	pageHelper := helpers.NewPageHelper(templates.Files())

	userHandler := handlers.NewUserHandler(userRepo, authHelper, pageHelper, "", "", []string{})

	return userHandler, userRepo
}

func makeRequestToHandlerWithClaims(claims *helpers.AuthClaims, path string, handler func(http.ResponseWriter, *http.Request), req *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	res := httptest.NewRecorder()
	authHelper := helpers.NewAuthHelper("test.com", "secret", []string{adminEmail})

	if claims != nil {
		tokenCookie := authHelper.NewJWTCookieFromClaims(claims)
		req.AddCookie(tokenCookie)
	}

	authMiddleware := middlewares.NewAuthMiddleware("/auth", []string{}, authHelper)
	router.Use(authMiddleware.EnsureAuth)
	router.HandleFunc(path, handler)
	router.ServeHTTP(res, req)

	return res
}

func TestUserHandler_List(t *testing.T) {
	t.Parallel()
	t.Run("Return unauthorized if not admin", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		req := httptest.NewRequest(http.MethodGet, "/user/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/", userHandler.List, req)

		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	})

	t.Run("Return OK for admin users", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodGet, "/user/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	t.Parallel()
	t.Run("Return unauthorized if not admin", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/delete", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/delete", userHandler.Delete, req)

		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	})

	t.Run("Return OK for admin users", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// User is in the DB
		user, err := userRepo.UserWithEmail(regularUserEmail)
		assert.NoError(t, err)
		assert.Equal(t, regularUserEmail, user.Email)

		// Delete
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/delete", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/delete", userHandler.Delete, req)
		location, _ := res.Result().Location()

		assert.Equal(t, http.StatusFound, res.Result().StatusCode)
		assert.Contains(t, location.Path, "/user/")

		// User nolonger in database
		user, err = userRepo.UserWithEmail(regularUserEmail)
		assert.ErrorIs(t, err, repos.ErrUserNotFound)
		assert.Nil(t, user)
	})
}

func TestUserHandler_View(t *testing.T) {
	t.Parallel()

	t.Run("Admin can view other user", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/", userHandler.View, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})

	t.Run("User can view self", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/", userHandler.View, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})

	t.Run("Regular user cannot view other users", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/", adminEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/", userHandler.View, req)
		location, _ := res.Result().Location()

		// Redirect to user page
		assert.Equal(t, http.StatusFound, res.Result().StatusCode)
		assert.Contains(t, location.Path, regularUserEmail)
	})
}

func TestUserHandler_Enroll(t *testing.T) {
	t.Parallel()

	t.Run("Admin can enroll new users", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// Make sure user doesn't exist
		err := userRepo.DeleteUser(regularUserEmail)
		assert.NoError(t, err)

		// Enroll
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/enroll", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/enroll", userHandler.Enroll, req)
		cacheControl := res.Result().Header.Get("Cache-Control")

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
		assert.Equal(t, "no-store, no-cache, must-revalidate", cacheControl)
	})

	t.Run("User can enroll self", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// Make sure user doesn't exist
		err := userRepo.DeleteUser(regularUserEmail)
		assert.NoError(t, err)

		// Enroll
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/enroll", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/enroll", userHandler.Enroll, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})

	t.Run("Regular user cannot enroll other users", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// Enroll
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/enroll", adminEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/enroll", userHandler.Enroll, req)
		location, _ := res.Result().Location()

		// Redirect to user page
		assert.Equal(t, http.StatusFound, res.Result().StatusCode)
		assert.Contains(t, location.Path, regularUserEmail)
	})
}

func TestUserHandler_Renew(t *testing.T) {
	t.Parallel()

	t.Run("Admin can renew existing users", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// Renew
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/renew", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/renew", userHandler.Renew, req)
		cacheControl := res.Result().Header.Get("Cache-Control")

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
		assert.Equal(t, "no-store, no-cache, must-revalidate", cacheControl)
	})

	t.Run("Admin cannot renew none-existing users", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// Make sure user doesn't exist
		err := userRepo.DeleteUser(regularUserEmail)
		assert.NoError(t, err)

		// Renew
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/renew", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/renew", userHandler.Renew, req)

		assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)
	})

	t.Run("User can enroll self", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// Renew
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/renew", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/renew", userHandler.Renew, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})

	t.Run("Regular user cannot enroll other users", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// Renew
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/renew", adminEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/renew", userHandler.Renew, req)
		location, _ := res.Result().Location()

		// Redirect to user page
		assert.Equal(t, http.StatusFound, res.Result().StatusCode)
		assert.Contains(t, location.Path, regularUserEmail)
	})
}
