package main

import (
	"log"
	"os"
	"testing"

	"webapp/pkg/repository/dbrepo"
)

var app application

func TestMain(m *testing.M) {
	pathToTemplates = "./../../templates/"
	app.Session = getSession()
	app.DSN = "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5"

	conn, err := app.connectToDb()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app.DB = &dbrepo.PostgresDbRepo{DB: conn}

	os.Exit(m.Run())
}
