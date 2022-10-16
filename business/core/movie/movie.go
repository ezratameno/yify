package movie

import (
	"fmt"

	"github.com/ezratameno/yify/business/data/store/movie"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Core manages the set of API's for user access.
type Core struct {
	log   *logrus.Entry
	movie movie.Store
}

// NewCore constructs a core for user api access.
func NewCore(log *logrus.Entry, db *sqlx.DB) Core {
	return Core{
		log:   log,
		movie: movie.NewStore(log, db),
	}
}

// Create inserts a new user into the database.
func (c Core) Create(nm movie.NewMovie) (movie.NewMovie, error) {
	// PERFORM PRE BUSINESS OPERATIONS
	c.log.WithFields(logrus.Fields{
		"status": fmt.Sprintf("trying to add %s", nm.Name),
	}).Info("create")
	newMovie, err := c.movie.Create(nm)
	if err != nil {
		return movie.NewMovie{}, fmt.Errorf("create: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS
	c.log.WithFields(logrus.Fields{
		"status": fmt.Sprintf("successfully added %s", nm.Name),
	}).Info("create")
	return newMovie, nil
}

// Create inserts a new user into the database.
func (c Core) GetMovies() ([]movie.Movie, error) {
	// PERFORM PRE BUSINESS OPERATIONS
	c.log.WithFields(logrus.Fields{
		"status": "trying to get movies",
	}).Info("get movies")
	movies, err := c.movie.GetMovies()
	if err != nil {
		c.log.WithFields(logrus.Fields{
			"ERROR": err,
		}).Error("get movies")
		return nil, err
	}
	c.log.WithFields(logrus.Fields{
		"status": "successfully got movies",
	}).Info("get movies")
	return movies, nil
}
