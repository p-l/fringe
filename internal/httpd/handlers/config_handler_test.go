package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/system"
	"github.com/stretchr/testify/assert"
)

func TestConfigHandler_Root(t *testing.T) {
	t.Parallel()

	t.Run("Return json format config with valid header", func(t *testing.T) {
		t.Parallel()

		googleConfig := system.GoogleConfig{
			ClientID:     "client-id",
			ClientSecret: "sercret",
		}
		configHandler := handlers.NewConfigHandler(googleConfig)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/", configHandler.Root)
		router.ServeHTTP(res, req)

		body := res.Body.String()

		assert.Contains(t, body, "\"google_client_id\":\"client-id\"")
		assert.Equal(t, res.Result().Header.Get("Content-Type"), "application/json")
		assert.Equal(t, res.Result().StatusCode, http.StatusOK)
	})
}
