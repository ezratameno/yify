// Package web contains a small web framework extension.
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

// A Handler is a type that handlers an http request within our own little mini
// framework
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// object for each of our http handlers.
type App struct {
	// Embedding - now App has all the methods of httptreemux.ContextMux
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}

}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// Handle sets a handler function for the given HTTP method and path pair
// to the application server mux
// group - the api group e.g v1
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {

	// First warp handler specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// The function to execute for each request
	h := func(w http.ResponseWriter, r *http.Request) {

		// Pull the context from the request.
		ctx := r.Context()

		// Set the context with the required values to
		// process the request.
		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}
		ctx = context.WithValue(ctx, key, &v)

		// Call the wrapped handler functions

		if err := handler(ctx, w, r); err != nil {
			a.SignalShutdown()
			return
		}
	}

	filepath := path
	if group != "" {
		filepath = "/" + group + path
	}
	a.ContextMux.Handle(method, filepath, h)

}
