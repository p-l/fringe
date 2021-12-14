package helpers

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

const baseLayoutFilename = "layouts/base.gohtml"

type PageHelper struct {
	files fs.FS
}

func NewPageHelper(templateFiles fs.FS) *PageHelper {
	return &PageHelper{
		files: templateFiles,
	}
}

func (p *PageHelper) RenderPage(httpResponse io.Writer, filename string, data interface{}) error {
	view, err := template.ParseFS(p.files, filename)
	if err != nil {
		return fmt.Errorf("error with %s template: %w", filename, err)
	}

	base, err := view.ParseFS(p.files, baseLayoutFilename)
	if err != nil {
		return fmt.Errorf("error with base template: %w", err)
	}

	if err = base.Execute(httpResponse, data); err != nil {
		return fmt.Errorf("error merging template: %w", err)
	}

	return nil
}
