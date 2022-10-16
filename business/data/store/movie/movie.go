// Package movie contains user related CRUD functionality.
package movie

import (
	"encoding/json"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// Store manages the set of API's for user access.
type Store struct {
	log *logrus.Entry
	db  *sqlx.DB
}

// NewStore constructs a user store for api access.
func NewStore(log *logrus.Entry, db *sqlx.DB) Store {
	return Store{
		log: log,
		db:  db,
	}
}

// Create inserts a new movie into the database.
func (s Store) Create(nm NewMovie) (NewMovie, error) {
	// Write the SQL statement we want to execute. I've split it over two lines
	// for readability (which is why it's surrounded with backquotes instead
	// of normal double quotes).
	// ? - placeholder, helps avoid SQL injection attacks
	stmt := `INSERT INTO movies (name,description,year,pageUrl,imageUrl,downloadLinks,categories,id)
	VALUES($1, $2, $3, $4, $5, $6, $7, $8)`

	// Use the Exec() method on the embedded connection pool to execute the
	// statement. The first parameter is the SQL statement, followed by the
	// title, content and expiry values for the placeholder parameters. This
	// method returns a sql.Result object, which contains some basic
	// information about what happened when the statement was executed.
	// expires - add this to the date, like a week from now
	_, err := s.db.Exec(stmt, nm.Name, nm.Description, nm.Year, nm.PageUrl, nm.ImageUrl, nm.DownloadLinks, nm.Categories, nm.ID)
	if err != nil {
		return nm, err
	}

	return nm, nil
}

// GetMovies returns the movies.
func (s Store) GetMovies() ([]Movie, error) {
	// Write the SQL statement we want to execute.
	stmt := `SELECT * FROM movies`

	// Use the Query() method on the connection pool to execute our
	// SQL statement. This returns a sql.Rows result set containing the result of
	// our query.
	rows, err := s.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	// We defer rows.Close() to ensure the sql.Rows resultset is
	// always properly closed before the Latest() method returns. This defer
	// statement should come *after* you check for an error from the Query()
	// method. Otherwise, if Query() returns an error, you'll get a panic
	// trying to close a nil resultset.
	defer rows.Close()

	// Initialize an empty slice to hold the models.Snippets objects.
	movies := []Movie{}

	// Use rows.Next to iterate through the rows in the resultset. This
	// prepares the first (and then each subsequent) row to be acted on by the
	// rows.Scan() method. If iteration over all the rows completes then the
	// result set automatically closes itself and frees-up the underlying
	// database connection.
	for rows.Next() {
		movie := &Movie{}

		// Use rows.Scan() to copy the values from each field in the row to the
		// new Snippet object that we created. Again, the arguments to row.Scan()
		// must be pointers to the place you want to copy the data into, and the
		// number of arguments must be exactly the same as the number of
		// columns returned by your statement.
		var categories string
		var downloadLinks string
		err = rows.Scan(&movie.ID, &movie.Name, &movie.Description, &movie.Year,
			&movie.PageUrl, &movie.ImageUrl, &downloadLinks, &categories)
		if err != nil {
			return nil, err
		}

		//  convert values
		movie.Categories = strings.Split(categories, "/")
		err = json.Unmarshal([]byte(downloadLinks), &movie.DownloadLinks)
		if err != nil {
			return nil, err
		}
		// Append it to the slice of snippets.

		movies = append(movies, *movie)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume that a successful iteration was completed
	// over the whole resultset.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// If everything went OK then return the Snippets slice.
	return movies, nil
}
