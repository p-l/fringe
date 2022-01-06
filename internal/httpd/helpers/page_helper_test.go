package helpers_test

import (
	"net/http/httptest"
	"testing"

	"github.com/p-l/fringe/internal/httpd/handlers"
	"github.com/p-l/fringe/internal/httpd/helpers"
	"github.com/p-l/fringe/templates"
	"github.com/stretchr/testify/assert"
)

func TestPageHelper_RenderPage(t *testing.T) {
	t.Parallel()

	t.Run("Fails if file doesn't exist", func(t *testing.T) {
		t.Parallel()

		pageHelper := helpers.NewPageHelper(templates.Files())

		res := httptest.NewRecorder()
		err := pageHelper.RenderPage(res, "does_not_exist.gohtml", nil)

		assert.Error(t, err)
	})

	t.Run("Load and render a template", func(t *testing.T) {
		t.Parallel()

		pageHelper := helpers.NewPageHelper(templates.Files())

		res := httptest.NewRecorder()
		err := pageHelper.RenderPage(res, "default/404.gohtml", handlers.NotFoundTemplateData{Path: "/full/path"})

		assert.NoError(t, err)
		assert.Contains(t, res.Body.String(), "/full/path")
	})
}
