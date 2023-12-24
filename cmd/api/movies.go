package main

import (
	"errors"
	"fmt"
	"greenlight.mayuraandrew.tech/internal/data"
	"greenlight.mayuraandrew.tech/internal/validator"
	"net/http"
	"strconv"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// initialize a new Validator instance
	v := validator.New()

	//// Use the Check() method to execute our validation checks. This will add the
	//// provided key and error message to the errors map if the check does not evaluate
	//// to true. For example, in the first line here we "check that the title is not
	//// equal to the empty string". In the second, we "check that the length of the title
	//// is less than or equal to 500 bytes" and so on.
	//v.Check(input.Title != "", "title", "must be provided")
	//v.Check(len(input.Title) <= 500, "title", "must not br more than 500 bytes long")
	//
	//v.Check(input.Year != 0, "year", "must be provided")
	//v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	//v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	//
	//v.Check(input.Runtime != 0, "runtime", "must be provided")
	//v.Check(input.Runtime > 0, "runtime", "must be a positive integer")
	//
	//v.Check(input.Genres != nil, "genres", "must be provided")
	//v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	//v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")
	//// note that we're using the Unique helper in the line below to check that all
	//// values in the input.Genres slice are unique.
	//v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")
	//// use the Valid() method to see if any of the checks failed. If they did, then use
	//// the failedValidationResponse() helper to send a response to the client, passing
	//// in the v.Errors map.
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorRespone(w, r, err)
		return
	}

	// when sending a HTTP response, we want to include a Location header to let the
	// client know which URL they can find the newly-created resource at. We make an
	// empty http.Header map and then use the Set() method to add a new Location header,
	// interpolating the system-generated ID for our new movie in the URL.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// write a JSON response with a 201 Created status code, the movie data in the
	// response body, and the Location header.
	err = app.writeJSON(w, http.StatusCreated, envelop{"movie": movie}, headers)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}

// add a showMovieHandler for the "GET /v1/movies/:id" endpoint.
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)
	// otherwise, interpolate the movie ID in a placeholder response.
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract the movie ID from the ID

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// fetch the existing movie record from the database, sending a 404 Not Found
	// response to the client if we couldn't find a matching record.

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	// declare an input struct to hold the expected data from the client.

	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// read the JSON request body data into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres // don't need to dereference a  slice
	}

	// validate the updated movie record, sending the client a 422 Unprocessble Entity
	// response if any checks fail.
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// pass the updated movie record to our new Update() method
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	// if the request contains a X-Expected-Version header, verify that the movie,
	// version in the database matches the expected version specified in the header.
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("x-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	// write the updated movie record in a JSON response.
	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract the movie ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// delete the movie from the database, sending a 404 Not Found response to the client if there isn;t a matching record.
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	// return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelop{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}
