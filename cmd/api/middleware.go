package main

import (
	"fmt"
	"golang.org/x/time/rate"
	"net"
	"net/http"
	"sync"
	"time"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create a deferred function (which will always be run in the event of panic
		// as Go unwinds the stack).
		defer func() {
			// the use builtin recover function to check it there has been a panic or not.
			if err := recover(); err != nil {
				// if there was a panic, set a "Connection: close" header on the response. This acts as a
				// trigger to make Go's HTTP server
				// automatically close the current connection after a response has been sent.

				w.Header().Set("Connection", "close")
				// the value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send client a 500
				// Internal Server Error response.
				app.serverErrorRespone(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// define a client struct to hold the rate limiter and last seen  time for each
	// client.

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// declare a mutex  and a map to hold the client's IP addresses and rate limiters.
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// launch a background goroutine which removes old entries form the clients map onnce
	// every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			// lock the mutex to prevent any rate limier checks from happening while
			// the cleanup is taking place.
			mu.Lock()

			// loop through all clients. if they haven't been seen within the last three minutes, delete the
			//corresponsing entry from the map.

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Importantly, unlock the mutex when the cleanup is complete.
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only carry out the check if rate limiting is enable.
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorRespone(w, r, err)
				return
			}

			mu.Lock()

			if _, found := clients[ip]; !found {
				// create and add a new client struct to the map if it doesn't already exist.
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}

			// update the last seen time for the client.
			clients[ip].lastSeen = time.Now()

			// Initialize a new rate limiter which allows an average of 2 requests per second,
			// with a maximum of 4 requests in a single 'burst'.

			// the function we are returning is a closure, which "closes over" the limiter.
			// variable.

			// extract the client's IP address from the request.

			// lock the mutex to prevent this code from being executed concurrently.

			// check to see if the IP address exists in the map. if it doesn't, then
			// initialize a new rate limiter and the IP address and limiter to the map.
			// call the allow() method to the rate limiter for the current IP address. If
			// the request isn't allowed, unlock the mutex and send a 429 too many requests.

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// very importantly, unlock the mutex before calling the next handler in the chain.
			// Notice that we DON'T use defer to unlock the mutex, as that would mean
			// that the mutex isn't unlocked until all the handlers downstream of this middleware have also returned.
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})

}
