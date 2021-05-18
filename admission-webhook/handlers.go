package main

import (
	"fmt"
	"net/http"
)

func (app *application) healthcheck(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Content-Type", "application/json")
	//Response message
	fmt.Fprintf(w, "%s", `{"msg": "server is healthy"}`)

}

func (app *application) validate(w http.ResponseWriter, r *http.Request) {
	//Set headers
	w.Header().Set("Content-Type", "application/json")
	//Response message
	fmt.Fprintf(w, "%s", `{"msg": "Validating endpoint"}`)

}
