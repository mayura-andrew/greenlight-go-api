package data

import (
	"database/sql"
	"errors"
)

// define a custom errors
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

//type Models struct {
//// Set the Movies field to be an interface containing the methods that both the
//// 'real' model and mock model need to support.
//Movies interface {
//Insert(movie *Movie) error
//Get(id int64) (*Movie, error)
//Update(movie *Movie) error
//Delete(id int64) error
//}
//}
//...
//// Create a helper function which returns a Models instance containing the mock models
//// only.
//func NewMockModels() Models {
//	return Models{
//		Movies: MockMovieModel{},
//	}
//}

// this Models struct wraps the MovieModel
type Models struct {
	Movies MovieModel
}

// for each of use, we also add a new() method which return a Models struct containing
// the initialized MovieModel.

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
