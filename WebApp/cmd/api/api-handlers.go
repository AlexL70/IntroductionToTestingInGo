package main

import (
	"errors"
	"net/http"
)

type Credentials struct {
	UserName string `json:"email"`
	Password string `json:"password"`
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	//	read a json payload
	err := app.readJSON(w, r, &creds)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized"), http.StatusBadRequest)
		return
	}

	//	look up the user by email address
	user, err := app.DB.GetUserByEmail(creds.UserName)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
		return
	}

	//	check password
	ok, err := user.PasswordMatches(creds.Password)
	if !ok || err != nil {
		app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	//	generate tokens (if password matches)
	pair, err := app.generateTokenPair(user)
	if err != nil {
		app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	// 	send token to user
	_ = app.writeJSON(w, http.StatusOK, pair)
}

func (app *application) refresh(w http.ResponseWriter, r *http.Request) {

}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {

}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) insertUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {

}

func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {

}
