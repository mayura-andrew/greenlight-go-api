Go Lang API project - Let's Go Further by Alex Edwards 


# API Endpoints

## Health Check

- `GET /v1/healthcheck`: Check the health of the application.

## Movies

- `GET /v1/movies`: List all movies. Requires `movies:read` permission.
- `POST /v1/movies`: Create a new movie. Requires `movies:write` permission.
- `GET /v1/movies/:id`: Retrieve a specific movie by its ID. Requires `movies:read` permission.
- `PATCH /v1/movies/:id`: Update a specific movie by its ID. Requires `movies:write` permission.
- `DELETE /v1/movies/:id`: Delete a specific movie by its ID. Requires `movies:write` permission.

## Users

- `POST /v1/users`: Register a new user.
- `PUT /v1/users/activated`: Activate a user.

## Authentication

- `POST /v1/tokens/authentication`: Create an authentication token.



## Account Creation and Activation

Before you can fully use the API, you need to create an account and activate it.

1. Create an account by sending a `POST` request to `/v1/users`. You will receive a hash code in the response.

2. Activate your account by sending a `PUT` request to `/v1/users/activated` with the hash code. Until your account is activated, you will only be able to read movies (`GET /v1/movies` and `GET /v1/movies/:id`).

## Authentication

After your account is activated, you need to authenticate to receive a Bearer token. This token is required to perform other tasks.

1. Authenticate by sending a `POST` request to `/v1/tokens/authentication`. You will receive a Bearer token in the response.

2. Include this Bearer token in the `Authorization` header of your requests to perform tasks that require permissions. For example, to create a new movie, send a `POST` request to `/v1/movies` with the Bearer token in the `Authorization` header.


## PostgreSQL Database

This application uses a PostgreSQL database to store data. 

### Local Development

For local development, you need to have PostgreSQL installed and running on your machine. Here are the steps to set it up:

1. Install PostgreSQL: You can download it from the official [PostgreSQL website](https://www.postgresql.org/download/).

2. Create a database for the application: You can do this through the PostgreSQL command line or a GUI tool like pgAdmin.

3. Set the database connection details as environment variables in your development environment. The application expects the following environment variables:
   - `DB_HOST`: The host of your PostgreSQL server.
   - `DB_PORT`: The port your PostgreSQL server is running on.
   - `DB_USER`: The username for your PostgreSQL database.
   - `DB_PASSWORD`: The password for your PostgreSQL database.
   - `DB_NAME`: The name of your PostgreSQL database.

### Production

In a production environment, you might use a managed PostgreSQL service. The setup will depend on your provider, but you will still need to set the same environment variables as in the local development setup.