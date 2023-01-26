package mid

import (
	"context"
	"net/http"

	"github.com/maxkulish/service-api/busines/sys/metrics"
	"github.com/maxkulish/service-api/foundation/web"
)

// Metrics updates program counters
func Metrics() web.Middleware {

	m := func(handler web.Handler) web.Handler {

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// Add the metrics into the context for metric gathering
			ctx = metrics.Set(ctx)

			// Call the next handler
			err := handler(ctx, w, r)

			// Handle updating the metrics that can be handled here

			// Increment the request and goroutines counter
			metrics.AddRequests(ctx)
			metrics.AddGoroutines(ctx)

			// Increment if there is an error flowing through the request
			if err != nil {
				metrics.AddErrors(ctx)
			}

			// Return the error so it can be handled further up the chain
			return err
		}

		return h
	}

	return m
}
