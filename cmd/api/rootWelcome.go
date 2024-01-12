package main

import (
	"fmt"
	"net/http"
)

// !Important --> this heathcheckHandler is implemented as a method on application struct
func (app *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Go Lang API Project</title>
	</head>
	<body>
		<h1>Go Lang API project - Let's Go Further by Alex Edwards</h1>
		<h2>API Endpoints</h2>
		<h3>Health Check</h3>
		<p>GET /v1/healthcheck: Check the health of the application.</p>
		<h3>Movies</h3>
		<p>GET /v1/movies: List all movies. Requires movies:read permission.</p>
		<p>POST /v1/movies: Create a new movie. Requires movies:write permission.</p>
		<p>GET /v1/movies/:id: Retrieve a specific movie by its ID. Requires movies:read permission.</p>
		<p>PATCH /v1/movies/:id: Update a specific movie by its ID. Requires movies:write permission.</p>
		<p>DELETE /v1/movies/:id: Delete a specific movie by its ID. Requires movies:write permission.</p>
		<h3>Users</h3>
		<p>POST /v1/users: Register a new user.</p>
		<p>PUT /v1/users/activated: Activate a user.</p>
		<h3>Authentication</h3>
		<p>POST /v1/tokens/authentication: Create an authentication token.</p>
		<h2>Account Creation and Activation</h2>
		<p>Before you can fully use the API, you need to create an account and activate it.</p>
		<p>Create an account by sending a POST request to /v1/users. You will receive a hash code in the response.</p>
		<p>Activate your account by sending a PUT request to /v1/users/activated with the hash code. Until your account is activated, you will only be able to read movies (GET /v1/movies and GET /v1/movies/:id).</p>
		<h2>Authentication</h2>
		<p>After your account is activated, you need to authenticate to receive a Bearer token. This token is required to perform other tasks.</p>
		<p>Authenticate by sending a POST request to /v1/tokens/authentication. You will receive a Bearer token in the response.</p>
		<p>Include this Bearer token in the Authorization header of your requests to perform tasks that require permissions. For example, to create a new movie, send a POST request to /v1/movies with the Bearer token in the Authorization header.</p>
		<h2>PostgreSQL Database</h2>
		<p>This application uses a PostgreSQL database to store data.</p>
		<h2>Local Development</h2>
		<p>For local development, you need to have PostgreSQL installed and running on your machine. Here are the steps to set it up:</p>
		<p>Install PostgreSQL: You can download it from the official PostgreSQL website.</p>
		<p>Create a database for the application: You can do this through the PostgreSQL command line or a GUI tool like pgAdmin.</p>
		<p>Set the database connection details as environment variables in your development environment. The application expects the following environment variables:</p>
		<p>DB_HOST: The host of your PostgreSQL server.</p>
		<p>DB_PORT: The port your PostgreSQL server is running on.</p>
		<p>DB_USER: The username for your PostgreSQL database.</p>
		<p>DB_PASSWORD: The password for your PostgreSQL database.</p>
		<p>DB_NAME: The name of your PostgreSQL database.</p>
		<p>Visit my website := https://mayuraandrew.tech</p>
	</body>
	</html>
    `)

}
