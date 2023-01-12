package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"webapp/pkg/data"
)

func Test_getTokenFromHeaderAndVerify(t *testing.T) {
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	var tests = []struct {
		name          string
		token         string
		errorExpected bool
		setHeader     bool
		issuer        string
	}{
		{"valid", fmt.Sprintf("Bearer %s", tokens.Token), false, true, app.Domain},
		{"expired", fmt.Sprintf("Bearer %s", expiredToken), true, true, app.Domain},
		{"no header", "", true, false, app.Domain},
		{"invalid token", fmt.Sprintf("Bearer %s500", tokens.Token), true, true, app.Domain},
		{"no bearer", fmt.Sprintf("Notbearer %s", tokens.Token), true, true, app.Domain},
		{"three parts", fmt.Sprintf("Bearer me %s", tokens.Token), true, true, app.Domain},
		//	make sure that the next test is the last one to run
		{"invalid issuer", fmt.Sprintf("Bearer %s", tokens.Token), true, true, "invalid.com"},
	}

	for _, e := range tests {
		if e.issuer != app.Domain {
			app.Domain = e.issuer
			tokens, _ = app.generateTokenPair(&testUser)
		}
		req, _ := http.NewRequest("GET", "/", nil)
		if e.setHeader {
			req.Header.Set("Authorization", e.token)
		}
		rr := httptest.NewRecorder()
		_, _, err := app.getTokenFromHeaderAndVerify(rr, req)
		if err != nil && !e.errorExpected {
			t.Errorf("%s: did not expect the error but got one - %q", e.name, err.Error())
		}

		if err == nil && e.errorExpected {
			t.Errorf("%s: expected error, but did not get one", e.name)
		}
		app.Domain = "example.com"
	}
}
