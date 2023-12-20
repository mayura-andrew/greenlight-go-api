package main

import (
	"fmt"
	"greenlight.mayuraandrew.tech/internal/data"
	"net/http"
	"time"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "create a new movie\n")
}

// add a showMovieHandler for the "GET /v1/movies/:id" endpoint.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// a new instance of the Movie struct

	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)
	// otherwise, interpolate the movie ID in a placeholder response.
	if err != nil {
		app.logger.Println(err)
		app.serverErrorRespone(w, r, err)
	}
}
