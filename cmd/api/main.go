package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// the application version nu,ber
const version = "1.0.0"

// the configuration settings for application
type config struct {
	port int
	env  string
}

// an application struct to hold the dependencies for HTTP handlers, helpers, and middleware.
type application struct {
	config config
	logger *log.Logger
}

// the main function code
func main() {
	// declare an instance of the config struct
	var cfg config

	// read the value of the port and env command-line flags into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// declare an instance of the application struct, containing the config struct and the logger.
	app := &application{
		config: cfg,
		logger: logger,
	}

	// declare an HTTP server with some sensible timeout settings, which listens on the port provided in the config struct

	// Use the httprouter instance returned by app.routes() as the server handler.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// start the HTTP server
	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)
}
