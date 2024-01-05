package handlers

import (
	"github.com/maxkulish/service-api/app/services/sales-api/handlers/v1/testgrp"
	"net/http"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"go.uber.org/zap"
)

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
}

// APIMux constructs an http.Handler with all application routes defined.
func APIMux(cfg APIMuxConfig) http.Handler {
	mux := httptreemux.NewContextMux()

	mux.Handle(http.MethodGet, "/test", testgrp.Test)

	return mux
}
