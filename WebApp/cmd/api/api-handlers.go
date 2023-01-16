package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
	"webapp/pkg/data"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
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

	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    pair.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	})

	// 	send token to user
	_ = app.writeJSON(w, http.StatusOK, pair)
}

func (app *application) refresh(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	refreshToken := r.Form.Get("refresh_token")
	claims := &Claims{}

	_, err = jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(app.JWTSecret), nil
	})
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	if time.Unix(claims.ExpiresAt.Unix(), 0).Sub(time.Now()) > 30*time.Second {
		app.errorJSON(w, errors.New("Refresh token does not need renewed yet"), http.StatusTooEarly)
		return
	}

	//	get the user id from the claims
	userID, err := strconv.Atoi(claims.Subject)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUser(userID)
	if err != nil {
		app.errorJSON(w, errors.New("unknown user"), http.StatusNotFound)
		log.Println(err)
		return
	}

	tokenPairs, err := app.generateTokenPair(user)
	if err != nil {
		app.errorJSON(w, errors.New("error generating token pair"), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    tokenPairs.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	})

	_ = app.writeJSON(w, http.StatusOK, tokenPairs)
}

func (app *application) AllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.DB.AllUsers()
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, users)
}

func (app *application) getUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	user, err := app.DB.GetUser(userID)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, user)
}

func (app *application) insertUser(w http.ResponseWriter, r *http.Request) {
	var user data.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	_, err = app.DB.InsertUser(user)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) updateUser(w http.ResponseWriter, r *http.Request) {
	var user data.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	err = app.DB.UpdateUser(user)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	err = app.DB.DeleteUser(userID)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
