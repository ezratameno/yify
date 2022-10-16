// Package usergrp maintains the group of handlers for movie access.
package moviegrp

import (
	"context"
	"net/http"

	"github.com/ezratameno/yify/business/core/movie"
	"github.com/ezratameno/yify/foundation/web"
)

// Handlers manages the set of user endpoints.
type Handlers struct {
	Movie movie.Core
}

// GetMovies returns a list of movies.
func (h Handlers) GetMovies(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	movies, err := h.Movie.GetMovies()
	if err != nil {
		return err
	}

	return web.Respond(context.Background(), w, movies, http.StatusOK)
}
