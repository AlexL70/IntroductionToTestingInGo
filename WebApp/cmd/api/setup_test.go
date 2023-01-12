package main

import (
	"os"
	"testing"
	"webapp/pkg/repository/dbrepo"
)

var app application

func TestMain(m *testing.M) {
	app.DB = &dbrepo.TestDbRepo{}
	app.Domain = "example.com"
	app.JWTSecret = "482ebbb8-6b80-44ef-8f05-7db3084a5fc726b0d372-56c5-4eb0-969a-951cfa13a9fc"
	os.Exit(m.Run())
}
