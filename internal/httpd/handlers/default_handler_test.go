package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/stretchr/testify/assert"
)

func TestDefaultHandler_Root(t *testing.T) {
	t.Parallel()

	t.Run("Not found returns a 404 status code", func(t *testing.T) {
		t.Parallel()

		defaultHandler := handlers.NewDefaultHandler()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		router := mux.NewRouter()
		router.NotFoundHandler = http.HandlerFunc(defaultHandler.NotFound)
		router.ServeHTTP(res, req)

		assert.Equal(t, http.StatusNotFound, res.Result().StatusCode)
	})
}
