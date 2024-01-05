package main

import (
	"errors"
	"greenlight.mayuraandrew.tech/internal/data"
	"greenlight.mayuraandrew.tech/internal/validator"
	"net/http"
	"time"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the email and password from the request body.
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// validate the email and password provided by the client.
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Lookup the user record on the email address. If no matching user was
	// found, then we call the app.invalidCredentialsResponse() helper to send a 401
	// Unauthorized response to the client.
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	// check if the provided password matches  the actual password for the user.
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorRespone(w, r, err)
		return
	}

	// if the password don't match, then we call the app.invalidCredentialsResponse()
	// helper again and return

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// otherwise, if the password is correct, we generate a new token with 24-hour
	//expiry time and the scope "authentication".
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorRespone(w, r, err)
		return
	}

	//encode the token to JSON and send it in the response  along with a 201 Created
	//status code.

	err = app.writeJSON(w, http.StatusCreated, envelop{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}
