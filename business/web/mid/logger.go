package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/ezratameno/yify/foundation/web"
	"github.com/sirupsen/logrus"
)

// Logger ...
func Logger(log *logrus.Entry) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {
		// Create the handler that will be attached in the middleware chain.

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return err
			}
			log.WithFields(logrus.Fields{
				"traceid":    v.TraceID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remoteaddr": r.RemoteAddr,
			}).Info("request started")

			// Call the next handler
			err = handler(ctx, w, r)

			log.WithFields(logrus.Fields{
				"traceid":    v.TraceID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"remoteaddr": r.RemoteAddr,
				"statuscode": v.StatusCode,
				"since":      time.Since(v.Now),
			}).Info("request completed")

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
