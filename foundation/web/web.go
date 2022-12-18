// Package web contains a small web framework extension
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

// A Handler is a type that handles an http request within our
// own little mini framework
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into out application and what configures our
// context object for each of our http handlers. Feel free to add any
// configuration data/logic on this App struct
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handle a set of routes for the application
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *App) Handle(method, group, path string, handler Handler, mw ...Middleware) {

	// First wrap handler specific middleware around this handler
	handler = wrapMiddleware(mw, handler)

	// Add the applications's general middleware to the handler chain
	handler = wrapMiddleware(a.mw, handler)

	// The function to execute fo each request
	h := func(w http.ResponseWriter, r *http.Request) {
		// Pre code processing

		if err := handler(r.Context(), w, r); err != nil {
			// Error handling
			return
		}

		// Post code processing
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	a.ContextMux.Handle(method, finalPath, h)
}
