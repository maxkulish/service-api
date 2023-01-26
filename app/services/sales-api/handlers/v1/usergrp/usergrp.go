// Package usergrp contains the set of handlers for the user resource.
package usergrp

import (
	"github.com/maxkulish/service-api/busines/core/user"
	"github.com/maxkulish/service-api/busines/sys/auth"
)

type Handlers struct {
	User user.Core
	Auth *auth.Auth
}
