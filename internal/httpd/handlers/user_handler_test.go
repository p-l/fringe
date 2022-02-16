package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/jaswdr/faker"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/internal/httpd/middlewares"
	"github.com/p-l/fringe/internal/mocks"
	"github.com/p-l/fringe/internal/repos"
	"github.com/stretchr/testify/assert"
)

const (
	adminEmail       = "admin@test.com"
	regularUserEmail = "user@test.com"
)

func createUserHandler(t *testing.T) (*handlers.UserHandler, *repos.UserRepository) {
	t.Helper()

	fake := faker.New()
	userRepo := mocks.NewMockUserRepository(t)

	for _, email := range []string{adminEmail, regularUserEmail} {
		_, err := userRepo.Create(email, fake.Person().Name(), fake.Internet().URL(), fake.Internet().Password())
		if err != nil {
			t.Fatalf("Could not add admin user to test database: %v", err)
		}
	}

	authHelper := helpers.NewAuthHelper("test.com", "secret", []string{})

	userHandler := handlers.NewUserHandler(userRepo, authHelper)

	return userHandler, userRepo
}

func makeRequestToHandlerWithClaims(claims *helpers.AuthClaims, path string, handler func(http.ResponseWriter, *http.Request), req *http.Request) *httptest.ResponseRecorder {
	router := mux.NewRouter()
	res := httptest.NewRecorder()
	authHelper := helpers.NewAuthHelper("test.com", "secret", []string{adminEmail})

	if claims != nil {
		token := authHelper.NewJWTSignedString(claims)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	authMiddleware := middlewares.NewAuthMiddleware("/auth", []string{"/"}, []string{}, authHelper)
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

		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	})

	t.Run("Return OK for admin users", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodGet, "/users/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})

	t.Run("Defaults to page 0 if argument is not a number", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodGet, "/users/?page=rabbit", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var usersInResponse []handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &usersInResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, usersInResponse)
	})

	t.Run("Defaults to 10 per_page if argument is not a number", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodGet, "/users/?per_page=rabbit", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var usersInResponse []handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &usersInResponse)
		assert.NoError(t, err)
		assert.NotEmpty(t, usersInResponse)
	})

	t.Run("Return no users when asking for a page past max", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodGet, "/users/?per_page=100&page=2", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var usersInResponse []handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &usersInResponse)
		assert.NoError(t, err)
		assert.Empty(t, usersInResponse)
	})

	t.Run("Returns only users matching a query string", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		escapedAdminEmail := url.QueryEscape(adminEmail)
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/?search=%s", escapedAdminEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var usersInResponse []handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &usersInResponse)
		assert.NoError(t, err)
		assert.Len(t, usersInResponse, 1)
	})

	t.Run("Returns no users when nothing matches search string", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/?search=%s", "ThereAreNoUserMatchingThisString"), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.List, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var usersInResponse []handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &usersInResponse)
		assert.NoError(t, err)
		assert.Empty(t, usersInResponse)
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

	t.Run("Admin get 404 on unknown user query", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/", "unknown@unknown.unknown"), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/", userHandler.View, req)

		assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)
	})

	t.Run("User can view self with email", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s/", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/{email}/", userHandler.View, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)
	})

	t.Run("User can view self with 'me' as email", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, "/users/me/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/{email}/", userHandler.View, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var user handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.Empty(t, user.Password)
		assert.Equal(t, user.Email, claims.Email)
	})

	t.Run("User enroll when first querying their info", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       "new-user@newuser.com",
			Name:        "New User",
			Permissions: "",
			Picture:     "",
		}

		// View
		req := httptest.NewRequest(http.MethodGet, "/users/me/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/{email}/", userHandler.View, req)

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

		// Forbidden
		assert.Equal(t, http.StatusForbidden, res.Result().StatusCode)
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
		err := userRepo.Delete(regularUserEmail)
		assert.NoError(t, err)

		// Renew
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/user/%s/renew", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/renew", userHandler.Renew, req)

		assert.Equal(t, http.StatusInternalServerError, res.Result().StatusCode)
	})

	t.Run("User can renew own password with email", func(t *testing.T) {
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

	t.Run("User can renew own password with 'me' as email", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		// Renew
		req := httptest.NewRequest(http.MethodGet, "/user/me/renew", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}/renew", userHandler.Renew, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var user handlers.UserResponse
		err := json.Unmarshal(res.Body.Bytes(), &user)
		assert.NoError(t, err)
		assert.NotEmpty(t, user.Password)
		assert.Equal(t, user.Email, claims.Email)
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

		assert.Equal(t, http.StatusForbidden, res.Result().StatusCode)
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

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/user/%s", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}", userHandler.Delete, req)

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
		user, err := userRepo.FindByEmail(regularUserEmail)
		assert.NoError(t, err)
		assert.Equal(t, regularUserEmail, user.Email)

		// Delete
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/user/%s", regularUserEmail), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/user/{email}", userHandler.Delete, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		// User not in database
		user, err = userRepo.FindByEmail(regularUserEmail)
		assert.ErrorIs(t, err, repos.ErrUserNotFound)
		assert.Nil(t, user)
	})

	t.Run("Refuses invalid email", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// Delete
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/", "invalid@email"), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/{email}/", userHandler.Delete, req)

		assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode)
	})

	t.Run("Refuses delete on not existing email", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		// Delete
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/%s/", "i.am@not.in.database.com"), nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/{email}/", userHandler.Delete, req)

		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var response handlers.UserActionResponse
		err := json.Unmarshal(res.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, response.Result, "not_found")
		assert.Nil(t, response.User)
	})
}

func TestUserHandler_Create(t *testing.T) {
	t.Parallel()

	t.Run("Return unauthorized if not admin", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       regularUserEmail,
			Permissions: "",
		}

		userData := handlers.UserCreateRequest{
			Email: "new.user@email.com",
			Name:  "New User",
		}
		jsonBytes, err := json.Marshal(userData)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(jsonBytes))
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.Create, req)

		assert.Equal(t, http.StatusUnauthorized, res.Result().StatusCode)
	})

	t.Run("Return invalid request on absent data", func(t *testing.T) {
		t.Parallel()

		userHandler, _ := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		req := httptest.NewRequest(http.MethodPost, "/users/", nil)
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.Create, req)

		assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode)
	})

	t.Run("Refuses when user outside of allowed domain", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		userData := handlers.UserCreateRequest{
			Email: "new.user@not-in-test.com",
			Name:  "New User",
		}
		jsonBytes, err := json.Marshal(userData)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(jsonBytes))
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.Create, req)

		// User is NOT in database
		user, err := userRepo.FindByEmail(userData.Email)
		assert.Error(t, err, repos.ErrUserNotFound)
		assert.Nil(t, user)

		// Refuses request
		assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode)
	})

	t.Run("Refuses malformed emails", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		userData := handlers.UserCreateRequest{
			Email: "new.user@not-a-valid-email",
			Name:  "New User",
		}
		jsonBytes, err := json.Marshal(userData)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(jsonBytes))
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.Create, req)

		// User is NOT in database
		user, err := userRepo.FindByEmail(userData.Email)
		assert.Error(t, err, repos.ErrUserNotFound)
		assert.Nil(t, user)

		// Refuses request
		assert.Equal(t, http.StatusBadRequest, res.Result().StatusCode)
	})

	t.Run("Refuses creating existing users", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		userData := handlers.UserCreateRequest{
			Email: regularUserEmail,
			Name:  "New User",
		}
		jsonBytes, err := json.Marshal(userData)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(jsonBytes))
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.Create, req)

		// User is in database
		user, err := userRepo.FindByEmail(userData.Email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, user.Email, userData.Email)

		// Refuses request
		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var response handlers.UserActionResponse
		err = json.Unmarshal(res.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, response.Result, "exists")
		assert.Nil(t, response.User)
	})

	t.Run("Return OK for admin users", func(t *testing.T) {
		t.Parallel()

		userHandler, userRepo := createUserHandler(t)
		claims := helpers.AuthClaims{
			Email:       adminEmail,
			Permissions: "admin",
		}

		userData := handlers.UserCreateRequest{
			Email: "new.user@test.com",
			Name:  "New User",
		}
		jsonBytes, err := json.Marshal(userData)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/users/", bytes.NewBuffer(jsonBytes))
		res := makeRequestToHandlerWithClaims(&claims, "/users/", userHandler.Create, req)

		// User is in database
		user, err := userRepo.FindByEmail(userData.Email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, user.Email, userData.Email)
		assert.Equal(t, user.Name, userData.Name)
		assert.NotEmpty(t, user.PasswordHash)

		// Valid response
		assert.Equal(t, http.StatusOK, res.Result().StatusCode)

		var response handlers.UserActionResponse
		err = json.Unmarshal(res.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, response.Result, "success")
		assert.NotNil(t, response.User)
		assert.Equal(t, response.User.Email, userData.Email)
		assert.Equal(t, response.User.Name, userData.Name)
		assert.NotEmpty(t, response.User.Password)
	})

	t.Run("", func(t *testing.T) {
		t.Parallel()
	})
}
