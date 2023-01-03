package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_application_addIpToContext(t *testing.T) {
	tests := []struct {
		headerName  string
		headerValue string
		addr        string
		emptyAddr   bool
	}{
		{"", "", "", false},
		{"", "", "", true},
		{"X-Forwarded-For", "192.3.2.1", "", false},
		{"", "", "hello:world", false},
	}

	//	create a dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//	make sure that the value exists in the context
		val := r.Context().Value(userIpKey)
		if val == nil {
			t.Errorf("\"%s\" not present", userIpKey)
		}
		//	make sure we get a string back
		ip, ok := val.(string)
		if !ok {
			t.Errorf("\"%v\" is not a valid IP value", val)
		}
		t.Log(ip)
	})

	for _, e := range tests {
		handlerToTest := app.addIpToContext(nextHandler)
		req := httptest.NewRequest("GET", "http://testing", nil)
		if e.emptyAddr {
			e.addr = ""
		}
		if len(e.headerName) > 0 {
			req.Header.Add(e.headerName, e.headerValue)
		}
		if len(e.addr) > 0 {
			req.RemoteAddr = e.addr
		}

		handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	}
}

func Test_application_ipFromContext(t *testing.T) {
	tests := []string{
		"192.3.2.1",
		"18.3.2.1",
		"193.3.2.1:3238",
		"MyIpAddress",
	}

	for _, e := range tests {
		var ctx = context.Background()
		if len(e) > 0 {
			ctx = context.WithValue(ctx, userIpKey, e)
		}
		ip := app.ipFromContext(ctx)
		if e != ip {
			t.Errorf("Expected: %q, got %q", e, ip)
		}
		t.Log(ip)
	}
}

func Test_app_auth(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})

	var tests = []struct {
		name   string
		isAuth bool
	}{
		{"logger in", true},
		{"not logger in", false},
	}

	for _, e := range tests {
		handlerToTest := app.auth(nextHandler)
		req := httptest.NewRequest("GET", "/user/profile", nil)
		req = addContextAndSessionToRequest(req, app)
		if e.isAuth {
			app.Session.Put(req.Context(), "user", data.User{ID: 1})
		}
		rr := httptest.NewRecorder()
		handlerToTest.ServeHTTP(rr, req)
		if e.isAuth && rr.Code != http.StatusOK {
			t.Errorf("%s: expected status code of %d, but got %d", e.name, http.StatusOK, rr.Code)
		}
		if !e.isAuth && rr.Code != http.StatusTemporaryRedirect {
			t.Errorf("%s: expected status code of %d, but got %d", e.name, http.StatusTemporaryRedirect, rr.Code)
		}
	}
}
