package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
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

	var app application
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
	const forwardedKey contextKey = "X-Forwarded-For"
	tests := []string{
		"192.3.2.1",
		"18.3.2.1",
		"193.3.2.1:3238",
		"MyIpAddress",
	}

	for _, e := range tests {
		var app application
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
