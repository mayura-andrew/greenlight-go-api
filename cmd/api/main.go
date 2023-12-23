package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq" // note that this _ blank identifier used for to stop the Go
	"greenlight.mayuraandrew.tech/internal/data"
	"log"
	"net/http"
	"os"
	"time"
	// compiler complaining that the package isn't being used.
)

// the application version number
const version = "1.0.0"

// the configuration settings for application
type config struct {
	port int
	env  string
	db   struct {
		dsn          string // db struct for to hold the configuration settings for our database connection pool.
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// an application struct to hold the dependencies for HTTP handlers, helpers, and middleware.
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

// the main function code
func main() {
	// declare an instance of the config struct
	var cfg config

	// read the value of the port and env command-line flags into the config struct.
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// read the DSN value from the db-dsn command-line flag into the config struct.
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// read the connection pool settings from command-line flags into the config struct.

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	// initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// call the openDB() helper function to create the connection pool.

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main() function exits.

	defer db.Close()

	// message to say that the connection pool has been successfully established.
	logger.Printf("database connection pool established.")

	// declare an instance of the application struct, containing the config struct and the logger.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
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
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// the openDB() function returns a sql.DB connection pool.

func openDB(cfg config) (*sql.DB, error) {
	// use sql.Open() to create an empty connction pool, using, the DSN from the config struct
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in-use + idle) connections in the pool,
	// note that passing a value less than or equal to 0 will mena there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections in the pool.

	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)
	// create  a context with a 5 second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use PingContext() to establish a new connection to the database, passing in the
	// context we created above as a parameter. If the connection couldn't be
	// established successfully within the 5 second deadline, then this will
	// return an error.

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	// return the sql.DB connection pool.
	return db, nil

}
