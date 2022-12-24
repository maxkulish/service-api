package web

import "errors"

// shutdownError is a type used to help with the graceful shutdown of the service
type shutdownError struct {
	Message string
}

// NewShutdownError a factory function which returns an error
// that causes the framework to signal a graceful shutdown
func NewShutdownError(message string) error {
	return &shutdownError{message}
}

// Error implements the error interface
func (she *shutdownError) Error() string {
	return she.Message
}

// IsShutdown checks to see if the shutdown error is
// contained in the specified error value
func IsShutdown(err error) bool {
	var she *shutdownError
	return errors.As(err, &she)
}
