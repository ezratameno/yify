// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

import (
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/ezratameno/yify/app/services/yify/handlers/debug/checkgrp"
	"github.com/ezratameno/yify/app/services/yify/handlers/v1/moviegrp"
	"github.com/ezratameno/yify/business/core/movie"
	"github.com/ezratameno/yify/business/web/mid"
	"github.com/ezratameno/yify/foundation/web"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux.
// Using the DefaultServerMux would be a security risk since a dependency could inject a handler
// into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}

// DebugMux registers all the debug standard library routes and the custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux.
// Using the DefaultServerMux would be a security risk since a dependency could inject a handler
// into our service without us knowing it.
func DebugMux(build string, log *logrus.Entry, db *sqlx.DB) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoint.
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
		DB:    db,
	}
	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux

}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *logrus.Entry
	DB       *sqlx.DB
}

// APIMux constructs an http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) *web.App {
	//  Construct the web.App which holds all routes.
	app := web.NewApp(
		cfg.Shutdown,
		// Inject the logger middleware
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Panics(),
	)

	// Load the routes for the different versions of the API.
	v1(app, cfg)
	return app
}

// v1 binds all the version 1 routes.
func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"

	mgh := moviegrp.Handlers{
		Movie: movie.NewCore(cfg.Log, cfg.DB),
	}

	app.Handle(http.MethodGet, version, "/movies", mgh.GetMovies)
}
