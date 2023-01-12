package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
