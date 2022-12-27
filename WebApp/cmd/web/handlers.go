package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
)

var pathToTemplates = "./templates/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{})
}

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *application) render(w http.ResponseWriter, r *http.Request, tmpl string, data *TemplateData) error {
	//	parse template
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, tmpl))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return err
	}

	data.IP = app.ipFromContext(r.Context())

	//	execute template
	err = parsedTemplate.Execute(w, data)
	if err != nil {
		return err
	}

	return nil
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Print(err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	//	validate data
	form := NewForm(r.PostForm)
	form.Required("email", "password")

	if !form.Valid() {
		fmt.Fprint(w, "failed validation")
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	log.Println(email, password)

	fmt.Fprint(w, email)
}
