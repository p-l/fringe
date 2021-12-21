package helpers

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"reflect"
	"time"
)

const baseLayoutTemplate = "layouts/base.gohtml"

type PageHelper struct {
	files fs.FS
}

func NewPageHelper(templateFiles fs.FS) *PageHelper {
	return &PageHelper{
		files: templateFiles,
	}
}

func (p *PageHelper) RenderPage(httpResponse io.Writer, templateFile string, data interface{}) error { // Create the last function
	templateBaseName := filepath.Base(templateFile)

	// Insert `last` function
	view := template.New(templateBaseName).Funcs(template.FuncMap{
		"last": func(x int, a interface{}) bool {
			return x == reflect.ValueOf(a).Len()-1
		},
		"unix_date_string": func(unix int64) string {
			return time.Unix(unix, 0).String()
		},
	})

	// Parse the template
	parsedView, err := view.ParseFS(p.files, templateFile)
	if err != nil {
		return fmt.Errorf("error with %s template: %w", templateFile, err)
	}

	// Add the parsed base layout
	base, err := parsedView.ParseFS(p.files, baseLayoutTemplate)
	if err != nil {
		return fmt.Errorf("error with default template: %w", err)
	}

	if err = base.Execute(httpResponse, data); err != nil {
		return fmt.Errorf("error merging template: %w", err)
	}

	return nil
}
