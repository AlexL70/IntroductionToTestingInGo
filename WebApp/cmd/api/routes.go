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
	//	mux.Use(app.enableCORS)

	//	authentication routes: auth handler and refresh handler
	mux.Post("/auth", app.authenticate)
	mux.Post("/refresh-token", app.refresh)

	//	test handler (to be able to ping API)
	mux.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		var payload = struct {
			Message string `json:"message"`
		}{
			Message: "Hello World!",
		}

		_ = app.writeJSON(w, http.StatusOK, payload)
	})

	//	protected routes
	mux.Route("/users", func(mux chi.Router) {
		//	use auth middleware

		mux.Get("/", app.AllUsers)
		mux.Get("/{userID}", app.getUser)
		mux.Delete("/{userID}", app.deleteUser)
		mux.Put("/", app.insertUser)
		mux.Patch("/{userID}", app.updateUser)
	})

	return mux
}
