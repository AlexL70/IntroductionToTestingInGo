package main

import (
	"html/template"
	"net/http"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{})
}

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *application) render(w http.ResponseWriter, r *http.Request, tmpl string, data *TemplateData) error {
	//	parse template
	parsedTemplate, err := template.ParseFiles("./templates/" + tmpl)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return err
	}
	//	execute template
	err = parsedTemplate.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}
