package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/maxkulish/service-api/business/sys/metrics"
	"github.com/maxkulish/service-api/foundation/web"
)

func Panics() web.Middleware {

	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function to recover from a panic
			// and set the err return variable after the fact.
			defer func() {
				if rec := recover(); rec != nil {
					// Stack trace will be provided
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE [%s]", rec, string(trace))

					// Updates the metrics stored in the context
					metrics.AddPanics(ctx)
				}
			}()

			// Call the next handler and set its return value in the err variable
			// In case this handler panics, defer will catch it
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
