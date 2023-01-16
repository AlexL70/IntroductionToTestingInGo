package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	//	register middleware
	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("./html/"))))

	mux.Route("/web", func(mux chi.Router) {
		mux.Post("/auth", app.authenticate)
		//	/refresh-token
		//	/logout
	})

	//	authentication routes: auth handler and refresh handler
	mux.Post("/auth", app.authenticate)
	mux.Post("/refresh-token", app.refresh)

	//	protected routes
	mux.Route("/users", func(mux chi.Router) {
		mux.Use(app.authRequired)

		mux.Get("/", app.AllUsers)
		mux.Get("/{userID}", app.getUser)
		mux.Delete("/{userID}", app.deleteUser)
		mux.Put("/", app.insertUser)
		mux.Patch("/", app.updateUser)
	})

	return mux
}
