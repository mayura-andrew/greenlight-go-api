package main

import (
	"errors"
	"greenlight.mayuraandrew.tech/internal/data"
	"greenlight.mayuraandrew.tech/internal/validator"
	"net/http"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Create an anonymous struct to hold the expected data from the request body.

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body into the anonymous struct.

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Copy the data from the request body into a new User struct. Notice also that we
	// set the Activated field to false, which isn't strictly necessary because the
	// Activated field will have the zero-value of false by default. But setting this
	// explicitly helps to make our intentions clear to anyone reading the code

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// use the Password.Set() method to generate and store the hashed and plaintext
	// password.
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorRespone(w, r, err)
		return
	}

	v := validator.New()

	// validate the user struct and return the error messages to the client messages to the client if any of the check fails.
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database.
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// if we get a ErrDuplicateEmail error, use the v.AddError() method to manually
		// add a message to the validator instance, and then call our
		// failedValidationResponse() helper
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	// call the Send() method our Mailer, passing in the user's email address,
	// name of the template file, and the User struct containing the new user's data.
	app.background(func() {

		// Send the welcome email
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
			return
		}
	})

	// write a JSON response containing the user data along with a 201 Created status
	// code.
	// Note that we also change this to send the client a 202 Accepted status code.
	// This status code indicates that the request has been accepted for processing, but
	// the processing has not been completed.
	err = app.writeJSON(w, http.StatusAccepted, envelop{"user": user}, nil)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}
