package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestForm_Has(t *testing.T) {
	form := NewForm(nil)
	has := form.Has("whatever")
	if has {
		t.Error("Form shows it has field when it should not")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	form = NewForm(postedData)
	has = form.Has("a")
	if !has {
		t.Error("Form shows it doesn't have field when it should")
	}
}

func TestForm_Required(t *testing.T) {
	r := httptest.NewRequest("POST", "http://whatever", nil)
	form := NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if form.Valid() {
		t.Error("Form shows valid when required fields are missing")
	}

	postedData := url.Values{}
	postedData.Add("a", "a")
	postedData.Add("b", "b")
	postedData.Add("c", "c")
	r, _ = http.NewRequest("POST", "http://whatever", nil)
	r.PostForm = postedData
	form = NewForm(r.PostForm)
	form.Required("a", "b", "c")
	if !form.Valid() {
		t.Error("Form shows invalid when all required fields are present")
	}
}
