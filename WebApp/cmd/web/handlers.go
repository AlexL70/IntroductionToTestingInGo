package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path"
	"time"
	"webapp/pkg/data"
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
	IP    string
	Data  map[string]any
	Error string
	Flash string
	User  data.User
}

func (app *application) render(w http.ResponseWriter, r *http.Request, tmpl string, td *TemplateData) error {
	//	parse template
	parsedTemplate, err := template.ParseFiles(path.Join(pathToTemplates, tmpl), path.Join(pathToTemplates, "base_layout.gohtml"))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return err
	}

	td.IP = app.ipFromContext(r.Context())
	td.Error = app.Session.PopString(r.Context(), "error")
	td.Flash = app.Session.PopString(r.Context(), "flash")

	//	execute template
	err = parsedTemplate.Execute(w, td)
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

	//	authenticate the user
	//	if user is not authenticated then retirect back with error
	user, err := app.DB.GetUserByEmail(email)
	if err != nil || !app.authenticate(r, user, password) {
		if err != nil {
			log.Println(err)
		}
		app.Session.Put(r.Context(), "error", "Invalid login!")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	//	ir user is authenticated then prevent fixation attack
	_ = app.Session.RenewToken(r.Context())

	//	store success message in session
	app.Session.Put(r.Context(), "flash", "Successfully logged in!")
	//	redirect to some other page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

func (app *application) authenticate(r *http.Request, user *data.User, password string) bool {
	if valid, err := user.PasswordMatches(password); err != nil || !valid {
		if err != nil {
			log.Println(fmt.Errorf("Login error: %e", err))
		}
		return false
	}

	app.Session.Put(r.Context(), "user", user)

	return true
}
