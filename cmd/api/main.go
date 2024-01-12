package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq" // note that this _ blank identifier used for to stop the Go
	"greenlight.mayuraandrew.tech/internal/data"
	"greenlight.mayuraandrew.tech/internal/jsonlog"
	"greenlight.mayuraandrew.tech/internal/mailer"
	"greenlight.mayuraandrew.tech/internal/vcs"
	// compiler complaining that the package isn't being used.
)

// Make version a variable (rather than a constant) and set its value to vcs.Version().
var (
	version = vcs.Version()
)

// the application version number
//const version = "1.0.0"

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

	// add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable.disable rate limiting altogether.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

// an application struct to hold the dependencies for HTTP handlers, helpers, and middleware.
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
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

	//flag.StringVar(&cfg.db.dsn, "db-dsn", "", "PostgreSQL DSN")

	// read the connection pool settings from command-line flags into the config struct.

	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.

	envVarValue := os.Getenv("SMTPPORT")

	if envVarValue == "" {
		fmt.Println("Environment variable is not set.")
		return
	}

	// Convert the environment variable value to an integer
	intValue, err := strconv.Atoi(envVarValue)
	if err != nil {
		fmt.Println("Error converting environment variable to integer:", err)
		return
	}
	fmt.Printf("%d", intValue)

	smtpSender, err := url.QueryUnescape(os.Getenv("SMTPSENDER"))
	if err != nil {
		log.Fatalf("Failed to decore the SMTPSENDER: %v", err)
	}
	fmt.Printf("%s", smtpSender)
	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTPHOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", intValue, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTPUSERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTPPASS"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", smtpSender, "SMTP sender")

	// Use the flag.Func() function to process the -cors-trusted-origins command line
	// flag. In this we use the strings.Fields() function to split the flag value into a
	// slice based on whitespace characters and assign it to our config struct.
	// Importantly, if the -cors-trusted-origins flag is not present, contains the empty
	// string, or contains only whitespace, then strings.Fields() will return an empty
	// []string slice.
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")
	flag.Parse()

	// If the version flag value is true, then print out the version number and
	// immediately exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}
	// initialize a new logger which writes messages to the standard out stream,
	// prefixed with the current date and time.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// call the openDB() helper function to create the connection pool.

	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}

	// Defer a call to db.Close() so that the connection pool is closed before the main() function exits.

	defer db.Close()

	// message to say that the connection pool has been successfully established.
	logger.PrintInfo("database connection pool established.", nil)

	// Publish a new "version" variable in the expvar handler containing our application
	// version number (currently the constant "1.0.0").
	expvar.NewString("version").Set(version)

	// publish the number of active goroutines.
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	// publish the database connection pool statistics.
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))

	// publish the current unix  timestamp.
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	// declare an instance of the application struct, containing the config struct and the logger.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	logger.PrintFatal(err, nil)
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
