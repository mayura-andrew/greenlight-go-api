package main

import (
	"fmt"
	"net/http"
)

// declare a handler which write a
// plain-text respone with information about the
// application status, operating environment and version

// !Important --> this heathcheckHandler is implemented as a method on application struct
func (app *application) welcomeMessage(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, "<h1>Welcome to our website!</h1>")

}
