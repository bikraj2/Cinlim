package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/v1/healthCheck", app.healthCheckHandler)
	// CRUD for Movie
	router.HandlerFunc(http.MethodGet, "/v1/movie", app.requirePermission("movies:read", app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movie", app.requirePermission("movies:write", app.createMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movie/:id", app.requirePermission("movies:read", app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movie/:id", app.requirePermission("movies:write", app.updateHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movie/:id", app.requirePermission("movies:write", app.deleteHanlder))
	//CRUD For User
	router.HandlerFunc(http.MethodPost, "/v1/user", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/user/activate", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/user/authenticate", app.createAuthenticationTokenHandler)
	// Route for Checking metrics
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())
	// Return the httpRouter Instance
	standard := alice.New(app.metrics, app.recoverPanic, app.enableCORS, app.rateLimit, app.authenticate)
	return standard.Then(router)
}
