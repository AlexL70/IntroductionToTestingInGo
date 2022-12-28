package main

import (
	"flag"
	"log"
	"net/http"
	"webapp/pkg/db"

	"github.com/alexedwards/scs/v2"
)

type application struct {
	Session *scs.SessionManager
	DB      db.PostgresConn
	DSN     string
}

func main() {
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
	app.DB = db.PostgresConn{DB: conn}

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
