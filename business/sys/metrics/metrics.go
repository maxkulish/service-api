package metrics

import (
	"context"
	"expvar"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The expvar package is already based on a singleton
// for the different metrics that are registered with the package
// so there isn't much choice here
var m *metrics

// metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently thanks to expvar.
// No extra abstraction is required
type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

// ==============================================================================

// metrics will be supported through the context

// ctxKeyMetric represents the type of value for the context key
type ctxKeyMetric int

// key is how metric values are stored/retrieved
const key ctxKeyMetric = 1

// Set sets the metrics data into the context
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// AddGoroutines increment the goroutines metric by 1
func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		if v.requests.Value()%100 == 0 { // Add every 100 request
			v.goroutines.Add(1)
		}
	}
}

// AddGoroutines increment the request metric by 1
func AddRequests(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.goroutines.Add(1)
	}
}

// AddErrors increment the errors metric by 1
func AddErrors(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.errors.Add(1)
	}
}

// AddPanics increment the panics metric by 1
func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.panics.Add(1)
	}
}
