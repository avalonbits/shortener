package embed

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	echo "github.com/labstack/echo/v4"
)

//go:embed *.tmpl
var templateFiles embed.FS

type Template struct {
	views map[string]*template.Template
}

func (t *Template) Render(w io.Writer, vName string, data any, c echo.Context) error {
	view, ok := t.views[vName]
	if !ok {
		panic(fmt.Sprintf("invalid view name:: %q", vName))
	}
	err := view.Execute(w, data)
	if err != nil {
		panic(err)
	}
	return nil
}

func (t *Template) NewView(name, file string) error {
	if _, ok := t.views[name]; ok {
		return fmt.Errorf("view with name %q already registered.", name)
	}

	view := template.Must(template.ParseFS(templateFiles, file))
	t.views[name] = view
	return nil
}

func Templates() *Template {
	return &Template{
		views: map[string]*template.Template{},
	}
}
