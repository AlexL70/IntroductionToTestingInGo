package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
)

var pathToTemplates = "./templates/"

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	//	create map for template data
	var td = make(map[string]any)
	//	check if data already exist in the session
	if !app.Session.Exists(r.Context(), "test") {
		app.Session.Put(r.Context(), "test", fmt.Sprintf("Hit this page at %s", time.Now().UTC().String()))
	}
	//	pass data to the page
	msg := app.Session.GetString(r.Context(), "test")
	td["test"] = msg
	_ = app.render(w, r, "home.page.gohtml", &TemplateData{Data: td})
}

func (app *application) Profile(w http.ResponseWriter, r *http.Request) {
	_ = app.render(w, r, "profile.page.gohtml", &TemplateData{})
}

type TemplateData struct {
	IP   string
	Data map[string]any
}

func (app *application) render(w http.ResponseWriter, r *http.Request, tmpl string, data *TemplateData) error {
	//	parse template
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, tmpl), path.Join(pathToTemplates, "base_layout.gohtml"))
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
		//	redirect to the login page with error messages
		log.Println(err)
		app.Session.Put(r.Context(), "error", "Invalid login creds")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := app.DB.GetUserByEmail(email)
	if err != nil {
		log.Println(err)
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//	temporary code
	log.Println(password, user.FirstName)

	//	authenticate the user
	//	if user is not authenticated then retirect back with error

	//	ir user is authenticated then prevent fixation attack
	_ = app.Session.RenewToken(r.Context())

	//	store success message in session
	//	redirect to some other page
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}
