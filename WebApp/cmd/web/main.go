package main

import (
	"encoding/gob"
	"flag"
	"log"
	"net/http"
	"webapp/pkg/data"
	"webapp/pkg/repository"
	"webapp/pkg/repository/dbrepo"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	Session *scs.SessionManager
	DB      repository.DatabaseRepo
	DSN     string
}

func main() {
	gob.Register(data.User{})
	//	set up an app config
	app := application{}
	flag.StringVar(&app.DSN, "dsn",
		"host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable timezone=UTC connect_timeout=5",
		"Postgress connection string")
	flag.Parse()
	conn, err := app.connectToDb()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	app.DB = &dbrepo.PostgresDbRepo{DB: conn}

	//	get a session manager
	app.Session = getSession()

	//	print out a message
	log.Println("Starting server on port 8080...")

	//	start the server
	err = http.ListenAndServe(":8080", app.routes())
	if err != nil {
		log.Fatal(err)
	}
}
