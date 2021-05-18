package main

import (
	"github.com/gorilla/mux"
)

//Refers to our application struct from earlier, as that has the healtcheck method attached to it
func (app *application) routes() *mux.Router {

	router := mux.NewRouter()
	//Get requests that go to /healtcheck will be handled by the healcheck method in the handlers.go file
	router.HandleFunc("/healthcheck", app.healthcheck).Methods("GET")
	router.HandleFunc("/validate", app.validate).Methods("POST")

	return router
}
