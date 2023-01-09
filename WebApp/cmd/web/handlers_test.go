package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"webapp/pkg/data"
)

func Test_application_handlers(t *testing.T) {
	var theTests = []struct {
		name                    string
		url                     string
		expectedStatusCode      int
		expectedUrl             string
		expectedFirstStatusCode int
	}{
		{"home", "/", http.StatusOK, "/", http.StatusOK},
		{"404", "/non_existent", http.StatusNotFound, "/non_existent", http.StatusNotFound},
		{"profile", "/user/profile", http.StatusOK, "/", http.StatusTemporaryRedirect},
	}

	routes := app.routes()

	//	create a test server
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: tr,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, e := range theTests {
		resp, err := ts.Client().Get(ts.URL + e.url)
		if err != nil {
			t.Log(err)
			t.Fatal(err)
		}
		if resp.StatusCode != e.expectedStatusCode {
			t.Errorf("for \"%s\": expected status is %d, but got %d", e.name, e.expectedStatusCode, resp.StatusCode)
		}

		if resp.Request.URL.Path != e.expectedUrl {
			t.Errorf("%s: expected final URL of %q, but got %q", e.name, e.expectedUrl, resp.Request.URL.Path)
		}

		resp2, _ := client.Get(ts.URL + e.url)
		if resp2.StatusCode != e.expectedFirstStatusCode {
			t.Errorf("for \"%s\": expected first returned status code is %d, but got %d", e.name, e.expectedFirstStatusCode, resp2.StatusCode)
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

func TestApp_renderWithBadTemplate(t *testing.T) {
	//	set template path to a location with bad template
	var oldPathToTemplates = pathToTemplates
	pathToTemplates = "./testdata/"

	req, _ := http.NewRequest("GET", "/", nil)
	req = addContextAndSessionToRequest(req, app)
	rr := httptest.NewRecorder()

	err := app.render(rr, req, "bad.page.gohtml", &TemplateData{})
	if err == nil {
		t.Error("Expected error from the bad template, but did not get one")
	}
	//	restore proper path to templates
	pathToTemplates = oldPathToTemplates
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

func Test_application_Login(t *testing.T) {
	var tests = []struct {
		name               string
		postedData         url.Values
		expectedStatusCode int
		expectedLoc        string
	}{
		{
			name: "valid login",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/user/profile",
		},
		{
			name: "invalid form",
			postedData: url.Values{
				"email":    {""},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "user not found",
			postedData: url.Values{
				"email":    {"wrong_user@example.com"},
				"password": {"secret"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
		{
			name: "invalid login",
			postedData: url.Values{
				"email":    {"admin@example.com"},
				"password": {"wrong_pwd"},
			},
			expectedStatusCode: http.StatusSeeOther,
			expectedLoc:        "/",
		},
	}

	for _, e := range tests {
		req, _ := http.NewRequest("POST", "/login", strings.NewReader(e.postedData.Encode()))
		req = addContextAndSessionToRequest(req, app)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.Login)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: returned wrong status code; expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		actualLoc, err := rr.Result().Location()
		if err == nil {
			if actualLoc.String() != e.expectedLoc {
				t.Errorf("%s: wrong final location; expected %q, but got %q", e.name, e.expectedLoc, actualLoc.String())
			}
		} else {
			t.Errorf("%s: no location header set", e.name)
		}
	}
}

func Test_application_UploadFiles(t *testing.T) {
	//	set up pipes
	pr, pw := io.Pipe()

	//	create new writer, of type *io.Writer
	writer := multipart.NewWriter(pw)

	//	create a waitgroup, and add 1 to it
	wg := &sync.WaitGroup{}
	wg.Add(1)

	//	simulate uploaoding a file using a goroutine and our writer
	go simulatePngUpload("./testdata/img.png", writer, t, wg)

	//	read from the pipe that receives data
	request := httptest.NewRequest("POST", "/", pr)
	request.Header.Add("Content-Type", writer.FormDataContentType())

	//	call app.UploadFiles
	uploadedFiles, err := app.UploadFiles(request, "./testdata/uploads/")
	if err != nil {
		t.Error(err)
	}

	//	perform our tests
	destFilePath := fmt.Sprintf("./testdata/uploads/%s", uploadedFiles[0].OriginalFileName)
	if _, err := os.Stat(destFilePath); os.IsNotExist(err) {
		t.Error(fmt.Errorf("File \"img.png\" was not saved in \"./testdata/uploads\" directory. Error: %w", err))
	}

	//	clean up
	wg.Wait()
	err = os.Remove(destFilePath)
	if err != nil {
		log.Println(fmt.Errorf("Cleanup error: %w", err))
	}

	wg.Wait()
}

func simulatePngUpload(fileToUpload string, w *multipart.Writer, t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	defer w.Close()

	//	create the form data filed 'file' with value being filename
	part, err := w.CreateFormFile("file", path.Base(fileToUpload))
	if err != nil {
		t.Error(err)
	}

	//	open actual file
	f, err := os.Open(fileToUpload)
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	//	decode the image
	img, _, err := image.Decode(f)
	if err != nil {
		t.Error(fmt.Errorf("Error decoding image: %w", err))
	}

	//	write the image to io.Writer
	err = png.Encode(part, img)
	if err != nil {
		t.Error(fmt.Errorf("Error writing image to the part: %w", err))
	}
}

func Test_application_UploadProfilePic(t *testing.T) {
	uploadPath = "./testdata/uploads"
	filePath := "./testdata/img.png"

	//	specify a field name for a form
	fieldName := "file"

	//	create buffer to act as request's body
	body := new(bytes.Buffer)

	//	crete a new writer
	mw := multipart.NewWriter(body)

	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}

	w, err := mw.CreateFormFile(fieldName, filePath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = io.Copy(w, file); err != nil {
		t.Fatal(err)
	}
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req = addContextAndSessionToRequest(req, app)
	app.Session.Put(req.Context(), "user", data.User{ID: 1})
	req.Header.Add("Content-Type", mw.FormDataContentType())
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.UploadProfilePic)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Bad status code returned; expected %d, but got %d", http.StatusSeeOther, rr.Code)
	}

	//	cleanup
	_ = os.Remove(fmt.Sprintf("%s/img.png", uploadPath))
}
