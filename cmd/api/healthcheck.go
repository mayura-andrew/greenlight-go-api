package main

import (
	"fmt"
	"net/http"
)

// declare a handler which write a
// plain-text respone with information about the
// application status, operating environment and version

// !Important --> this heathcheckHandler is implemented as a method on application struct
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: available")
	fmt.Fprintf(w, "environment: %s\n", app.config.env)
	fmt.Fprintf(w, "version: %s\n", version)
}
