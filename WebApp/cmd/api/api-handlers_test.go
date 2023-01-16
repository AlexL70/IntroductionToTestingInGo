package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"webapp/pkg/data"

	"github.com/go-chi/chi/v5"
)

func Test_app_authenticate(t *testing.T) {
	var theTests = []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid user", `{"email":"admin@example.com", "password": "secret"}`, http.StatusOK},
		{"not json", `Badly formatted input`, http.StatusBadRequest},
		{"empty json", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email":"", "password": "secret"}`, http.StatusUnauthorized},
		{"empty password", `{"email":"admin@example.com"}`, http.StatusUnauthorized},
		{"invalid user", `{"email":"admin@secret.com", "password": "secret"}`, http.StatusUnauthorized},
		{"wrong password", `{"email": "admin@example.com", "password": "bad_pwd"}`, http.StatusUnauthorized},
	}

	for _, e := range theTests {
		var reader io.Reader
		reader = strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: returned wrong status code. Expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_app_refresh(t *testing.T) {
	var tests = []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid", "", http.StatusOK, true},
		{"too early to expire", "", http.StatusTooEarly, false},
		{"expired", expiredToken, http.StatusBadRequest, false},
	}

	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 30
			}
			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = e.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req, _ := http.NewRequest("POST", "/refresh_token", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.refresh)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status of %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}

		//	Set refresh token expiry time to its initial value!!!
		refreshTokenExpiry = oldRefreshTime
	}
}

func Test_app_userHandlers(t *testing.T) {
	var tests = []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"allUsers", "GET", "", "", app.AllUsers, http.StatusOK},
		{"deleteUser", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"deleteUser bad URL param", "DELETE", "", "Y", app.deleteUser, http.StatusBadRequest},
		{"getUser valid", "GET", "", "1", app.getUser, http.StatusOK},
		{"getUser valid bad URL param", "GET", "", "X", app.getUser, http.StatusBadRequest},
		{"getUser invalid", "GET", "", "2", app.getUser, http.StatusBadRequest},
		{"updateUser valid", "PATCH",
			`{
				"id":1, "first_name": "Administrator", "last_name": "User", "email":"admin@example.com"
			}`, "", app.updateUser, http.StatusNoContent},
		{"updateUser invalid", "PATCH",
			`{
				"id":10, "first_name": "Administrator", "last_name": "User", "email":"admin@example.com"
			}`, "", app.updateUser, http.StatusBadRequest},
		{"updateUser invalid json", "PATCH",
			`{
				"id":1, first_name: "Administrator", "last_name": "User", "email":"admin@example.com"
			}`, "", app.updateUser, http.StatusBadRequest},
		{"insertUser valid", "PUT",
			`{
				"first_name": "Jack", "last_name": "Smith", "email":"jack@example.com"
			}`, "", app.insertUser, http.StatusNoContent},
		{"insertUser invalid", "PUT",
			`{
				"foo": "bar",
				"first_name": "Jack", "last_name": "Smith", "email":"jack@example.com"
			}`, "", app.insertUser, http.StatusBadRequest},
		{"insertUser invalidjson", "PUT",
			`{
				"first_name": Jack, "last_name": "Smith", "email":"jack@example.com"
			}`, "", app.insertUser, http.StatusBadRequest},
	}

	for _, e := range tests {
		var req *http.Request
		if e.json == "" {
			req, _ = http.NewRequest(e.method, "/", nil)
		} else {
			req, _ = http.NewRequest(e.method, "/", strings.NewReader(e.json))
		}

		if e.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(e.handler)
		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf("%s: wrong status returned; expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}

func Test_app_refreshUsingCookie(t *testing.T) {
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	testCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	badCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    "some bad token",
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	var tests = []struct {
		name           string
		addCookie      bool
		cookie         *http.Cookie
		expectedStatus int
	}{
		{"valid cookie", true, testCookie, http.StatusOK},
		{"invalid cookie", true, badCookie, http.StatusBadRequest},
		{"no cookie", false, testCookie, http.StatusUnauthorized},
	}

	for _, e := range tests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		if e.addCookie {
			req.AddCookie(e.cookie)
		}
		handler := http.HandlerFunc(app.refreshUsingCookie)
		handler.ServeHTTP(rr, req)
		if e.expectedStatus != rr.Code {
			t.Errorf("%s: wrong status code returned; expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}

func Test_app_deleteRefreshCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logoun", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.deleteRefreshCookie)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("wrong status; expected %d, but god %d", http.StatusAccepted, rr.Code)
	}

	foundCookie := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "__Host-refresh_token" {
			foundCookie = true
			if c.Expires.After(time.Now()) {
				t.Errorf("cookie expiration date in future, and should not be: %v", c.Expires.UTC())
			}
			break
		}
	}
	if !foundCookie {
		t.Error("__Host-refresh_token cookie not found!")
	}
}
