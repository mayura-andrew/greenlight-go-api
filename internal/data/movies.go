package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"greenlight.mayuraandrew.tech/internal/validator"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres"`
	Version   int32     `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `INSERT INTO movies 
    (title, year, runtime, genres) 
	VALUES ($1, $2, $3, $4) 
	RETURNING id, created_at, version`

	// create an args slice containing the values for the placeholder parameters from
	// the movie struct. Declaring this slice immedialety next to our SQL query helps to
	// make it nice and clear *what values are beign used where* in the query.

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// passing in the args slice as a variadic parameter and scanning the system-
	// generated id, created_at, and version values into the movie struct.

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// define the SQL query for retrieving the movie data.

	query := `SELECT id, created_at, title, year, runtime, genres, version 
	FROM movies WHERE id = $1`

	// declare a Movie struct to hold the data returned by the query.
	var movie Movie

	// use the context.WithTimeout() function to create a context.Context which
	// carries a 3-second timeout deadline.  note that we're using the empty context.Background()
	// as the "parent" context.

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// use defer to make sure that we cancel context before the Get()
	// method returns.
	defer cancel()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version)

	// handle any errors. if there was no matching movie found,
	// Scan() will return a sql.ErrNoRows error.
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// otherwise, return a pointer to the Movie struct
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `UPDATE movies SET title = $1, year = $2, runtime = $3, genres =$4, version = version + 1 
              WHERE id = $5 AND version = $6 
              RETURNING version`

	// create an args slice containing the values for the placeholder parameters.
	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// scanning the new version value into the movie struct
	// execute the SQL query. If no matching row could be found, we know the movie,
	// version has changed (or the record has been deleted) and we return
	// our custom ErrEditConflict error.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	// Construct the SQL query to delete the record.
	query := `DELETE FROM movies WHERE id = $1`

	// execute the SQL query using Exec() method, passing in the id variables as the value for the placeholder parameter. The Exec() method
	// returns a sql.Result object.

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Call the RowsAffected() method on the sql.Result object to get the number of rows.
	// affected by the query.
	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	// if no rows were affected, we know that the movies table didn't contain a record
	// with the provided ID at the moment we tried to delete it. In that case we
	// return an ErrRecordNotFound error.

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// construct the SQL query to retrieve all movie reords.

	// full-text search for the title filter
	//
	query := fmt.Sprintf(`SELECT id, created_at, title, year, runtime, genres, version
	FROM movies 
	WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// create a context with a 3-second timeout.

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	/// as our SQL query now has quite a few placeholder parameters, let's collect the values
	// values for the placeholders in a slice. Notice here how we call the limit() and offset() methods
	// on the Filters struct to get the appropriate values for the
	// LIMIT and OFFSET clauses.
	args := []interface{}{title, pq.Array(genres), filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	// importantly, defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	// Initialize an empty slice to hold the movie data.
	movies := []*Movie{}

	// use the rows.Next to iterate through the rows in the resultset.

	for rows.Next() {
		// Initialize an empty Movie struct to hold the data for an individual movie.
		var movie Movie

		// scan the values from the row into the Movie struct.
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version)

		if err != nil {
			return nil, err
		}

		// Add the Movie struct to the slice.

		movies = append(movies, &movie)
	}

	// when the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// if everything went ok, then return the slice of movies.
	return movies, nil

}

//type MockMovieModel struct{}
//
////
//func (m MockMovieModel) Insert(movie *Movie) error{
//	// Mock the action
//
//}
//
//func (m MockMovieModel) Get(id int64) (*Movie, error) {
//	// Mock the action
//}

//func (m MockMovieModel) Update (movie *Movie) error {
//	// mock the action
//}

//func (m MockMovieModel) Get(id int64) (*Movie, error) {
//	// Mock the action
//}
