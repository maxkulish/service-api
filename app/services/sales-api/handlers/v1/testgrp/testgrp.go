// Package testgrp provides support for the testgrp API.
package testgrp

import (
	"context"
	"errors"
	"math/rand"
	"net/http"

	"github.com/maxkulish/service-api/busines/sys/validate"
	"github.com/maxkulish/service-api/foundation/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

	if n := rand.Intn(100); n%2 == 0 {
		// We should never see this error. Correct error is 500
		// return errors.New("untrusted error")
		// return web.NewShutdownError("restart service")
		// panic("testing panic")
		return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
	}
	status := struct {
		Status string
	}{
		Status: "OK",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
