package views

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)


type Template struct {
  htmlTpl *template.Template
}

func (t Template) Execute(w http.ResponseWriter, data interface{}) {
  w.Header().Set("Content-Type", "text/html; charset=utf-8")
  err := t.htmlTpl.Execute(w, data)
  if err != nil {
    fmt.Errorf("template execution: %v", err)
    http.Error(w, "There was an error executing the template", http.StatusInternalServerError)
    return
  }
}

func Parse(filepath string) (Template, error) {
  htmlTpl, err := template.ParseFiles(filepath)
  if err != nil {
    return Template{}, fmt.Errorf("template parsing: %v", err)
  }
  return Template{htmlTpl}, nil
}

func ParseFS(fs fs.FS, pattern string) (Template, error) {
  htmlTpl, err := template.ParseFS(fs, pattern)
  if err != nil {
    return Template{}, fmt.Errorf("template parsing: %v", err)
  }
  return Template{htmlTpl}, nil
}

func Must(t Template, err error) Template {
  if err != nil {
    panic(err)
  }
  return t
}
