package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	if app.Session.Exists(r.Context(), "user") {
		td.User = app.Session.Get(r.Context(), "user").(data.User)
	}

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

func (app *application) UploadProfilePic(w http.ResponseWriter, r *http.Request) {
	//	call a function that extracts file from an upload (request)
	files, err := app.UploadFiles(r, "./static/img")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//	get the user from the session
	user := app.Session.Get(r.Context(), "user").(data.User)

	//	create a var of type data.UserImage
	img := data.UserImage{
		UserID:   user.ID,
		FileName: files[0].OriginalFileName,
	}

	//	insert the user image record into user_images table
	_, err = app.DB.InsertUserImage(img)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	//	refresh the sessional variable "user"
	updatedUser, err := app.DB.GetUser(user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	app.Session.Put(r.Context(), "user", updatedUser)

	log.Println(updatedUser)
	//	redirect back to profile page
	http.Redirect(w, r, "/user/profile", http.StatusSeeOther)
}

type UploadedFile struct {
	OriginalFileName string
	FileSize         int64
}

func (app *application) UploadFiles(r *http.Request, uploadDir string) ([]*UploadedFile, error) {
	var uploadedFiles []*UploadedFile

	err := r.ParseMultipartForm(int64(1024 * 1024 * 5)) //	parse not more than 5 megabytes
	if err != nil {
		return nil, fmt.Errorf("error uploading file: %w", err)
	}

	for _, fHeaders := range r.MultipartForm.File {
		for _, hdr := range fHeaders {
			uploadedFiles, err = func(uploadedFiles []*UploadedFile) ([]*UploadedFile, error) {
				var uploadedFile UploadedFile
				infile, err := hdr.Open()
				if err != nil {
					return nil, err
				}
				defer infile.Close()

				uploadedFile.OriginalFileName = hdr.Filename
				var outfile *os.File
				defer outfile.Close()
				if outfile, err = os.Create(filepath.Join(uploadDir, uploadedFile.OriginalFileName)); err != nil {
					return nil, err
				}
				fileSize, err := io.Copy(outfile, infile)
				if err != nil {
					return nil, err
				}
				uploadedFile.FileSize = fileSize
				uploadedFiles = append(uploadedFiles, &uploadedFile)

				return uploadedFiles, nil
			}(uploadedFiles)
			if err != nil {
				return uploadedFiles, fmt.Errorf("error uploading files: %w", err)
			}
		}
	}

	return uploadedFiles, nil
}
