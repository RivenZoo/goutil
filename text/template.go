package text

import (
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
)

var (
	ErrTmplNotExist = errors.New("template not exist")
)

type TemplateRender struct {
	tmpls map[string]*template.Template
}

// load template file *.tmpl in dir tmplDir and render
func NewTemplateRender(tmplDir string) *TemplateRender {
	fileInfo, err := ioutil.ReadDir(tmplDir)
	if err != nil {
		panic(err)
	}
	files := make([]string, 0, len(fileInfo))
	for _, fi := range fileInfo {
		name := fi.Name()
		if strings.HasSuffix(name, ".tmpl") {
			files = append(files, filepath.Join(tmplDir, name))
		}
	}
	allTmpl, err := template.ParseFiles(files...)
	if err != nil {
		panic(err)
	}
	tmplRender := &TemplateRender{
		tmpls: make(map[string]*template.Template),
	}
	tmpls := allTmpl.Templates()
	for _, tmpl := range tmpls {
		tmplRender.tmpls[tmpl.Name()] = tmpl
	}
	return tmplRender
}

func (r *TemplateRender) Render(w io.Writer, tmplName string, data interface{}) error {
	t, ok := r.tmpls[tmplName]
	if !ok {
		return ErrTmplNotExist
	}
	return t.Execute(w, data)
}
