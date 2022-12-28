package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test_application_handlers(t *testing.T) {
	var theTests = []struct {
		name               string
		url                string
		expectedStatusCode int
	}{
		{"home", "/", http.StatusOK},
		{"404", "/non_existent", http.StatusNotFound},
	}

	routes := app.routes()

	//	create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	pathToTemplates = "./../../templates/"

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for \"%s\": expected status is %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}
	}
}

func TestAppHome(t *testing.T) {
	var tests = []struct {
		name         string
		putInSession string
		expectedHTML string
	}{
		{"first visit", "", "<small>From session:"},
		{"second visit", "some session info", "<small>From session: some session info"},
	}

	for _, e := range tests {
		//	create a request
		req, _ := http.NewRequest("GET", "/", nil)
		req = addContextAndSessionToRequest(req, app)

		//	put data to the session
		_ = app.Session.Destroy(req.Context())
		if e.putInSession != "" {
			app.Session.Put(req.Context(), "test", e.putInSession)
		}

		//	create a response writer
		rr := httptest.NewRecorder()
		//	create a handler
		handler := http.HandlerFunc(app.Home)
		handler.ServeHTTP(rr, req)
		//	check status code
		if rr.Code != http.StatusOK {
			t.Errorf("TestAppHome returned wrong status code; expected %d, but got %d", http.StatusOK, rr.Code)
		}
		body, _ := io.ReadAll(rr.Body)
		if !strings.Contains(string(body), e.expectedHTML) {
			t.Errorf("%q: did not find %q in response body", e.name, e.expectedHTML)
		}
	}
}

func getCtx(r *http.Request) context.Context {
	ctx := context.WithValue(r.Context(), userIpKey, "unknown")
	return ctx
}

func addContextAndSessionToRequest(r *http.Request, app application) *http.Request {
	r = r.WithContext(getCtx(r))
	ctx, _ := app.Session.Load(r.Context(), r.Header.Get("X-Session"))
	return r.WithContext(ctx)
}
