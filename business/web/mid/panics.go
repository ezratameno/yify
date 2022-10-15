package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/ezratameno/yify/foundation/web"
)

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handled in Errors.
func Panics() web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {
		// Create the handler that will be attached in the middleware chain.

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic and set the err return
			// variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {

					trace := debug.Stack()
					// Stack trace will we provided.
					err = fmt.Errorf("PANIC [%v] TRACE [%s]", rec, string(trace))

				}
			}()

			// Call the next handler and set its return value in the err variable.
			return handler(ctx, w, r)
		}
		return h
	}
	return m
}