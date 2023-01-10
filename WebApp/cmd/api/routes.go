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

	//	test handler (to be able to ping API)

	//	protected routes

	return mux
}
