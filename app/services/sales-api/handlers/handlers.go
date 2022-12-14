// Package handlers contains the full set
// of handler functions and routes supported by teh we api
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/maxkulish/service-api/app/services/sales-api/handlers/debug/checkgrp"
	"github.com/maxkulish/service-api/app/services/sales-api/handlers/v1/testgrp"
	"github.com/maxkulish/service-api/business/sys/auth"
	"github.com/maxkulish/service-api/business/web/mid"
	"github.com/maxkulish/service-api/foundation/web"
	"go.uber.org/zap"
)

// DebugStandardLibraryMux returns a mux that contains all the standard library
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk
// since a dependency could inject a handler into our service without us knowing it.
func DebugMux(build string, log *zap.SugaredLogger) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
	}

	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)
	return mux
}

// APIMuxConfig contains all the mandatory systems required by handlers
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	Auth     *auth.Auth
}

// APIMux constructs an http.Handler with all application routes defined
func APIMux(cfg APIMuxConfig) *web.App {
	// Construct the web.App which holds all routes
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(), // Panics should be always at the end of the slice
	)

	// Load the routes for the different versions of the API
	v1(app, cfg)

	return app
}

// v1 binds all the version 1 routes
func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"
	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}

	app.Handle(http.MethodGet, version, "/test", tgh.Test)
	app.Handle(
		http.MethodGet,
		version,
		"/test-auth",
		tgh.Test,
		mid.Authenticate(cfg.Auth),
		mid.Authorize("ADMIN"),
	)
}
