package main

import (
	"net/http"
)

// declare a handler which write a
// plain-text respone with information about the
// application status, operating environment and version

// !Important --> this heathcheckHandler is implemented as a method on application struct
func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a map which holds the information that we want to send in the response.

	env := envelop{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	// add a 4 second delay.
	//	time.Sleep(4 * time.Second)

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logger.PrintError(err, nil)
		app.serverErrorRespone(w, r, err)
	}
}
