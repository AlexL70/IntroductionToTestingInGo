package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"webapp/pkg/repository"
	"webapp/pkg/repository/dbrepo"
)

const port = 8090

type application struct {
	DSN       string
	DB        repository.DatabaseRepo
	Domain    string
	JWTSecret string
}

func main() {
	var app application
	flag.StringVar(&app.Domain, "domain", "example.com", "Domain for the application e.g. \"company.com\"")
	flag.StringVar(&app.DSN, "dsn",
		"host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5",
		"Postgress connection string")
	flag.StringVar(&app.JWTSecret, "jwn-secret", "482ebbb8-6b80-44ef-8f05-7db3084a5fc726b0d372-56c5-4eb0-969a-951cfa13a9fc", "signing secret")
	flag.Parse()

	conn, err := app.connectToDb()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app.DB = &dbrepo.PostgresDbRepo{DB: conn}

	log.Printf("Starting API on port %d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
