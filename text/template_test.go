package text

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func genTemplateDir() (string, error) {
	dir, err := ioutil.TempDir("", "tmpl_test")
	if err != nil {
		return "", err
	}

	tmpl1 := `{{ define "Test1" }}hello,{{ .Name }}{{ end }}`
	tmpl2 := `{{ define "Test2" }}<body>{{ template "Test1" .Data }}</body>{{ end }}`
	if err = ioutil.WriteFile(filepath.Join(dir, "1.tmpl"), []byte(tmpl1), 0666); err != nil {
		return "", err
	}
	if err = ioutil.WriteFile(filepath.Join(dir, "2.tmpl"), []byte(tmpl2), 0666); err != nil {
		return "", err
	}
	return dir, nil
}

func TestTemplateRender(t *testing.T) {
	dir, err := genTemplateDir()
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(dir)
	render := NewTemplateRender(dir)
	err = render.Render(os.Stdout, "Test1", map[string]string{
		"Name": "there",
	})
	if err != nil {
		t.FailNow()
	}
	err = render.Render(os.Stdout, "Test2", map[string]interface{}{
		"Data": map[string]string{"Name": "there"},
	})
	if err != nil {
		t.FailNow()
	}
}
