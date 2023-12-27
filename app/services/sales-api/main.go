package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/maxkulish/service-api/foundation/logger"
	"go.uber.org/zap"
)

var build = "develop"

func main() {
	log, err := logger.New("sales-api")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer log.Sync()

	if err := run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
		log.Sync()
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {
	// -------------------------------------------------
	// GOMAXPROCS
	// The argument 0 in runtime.GOMAXPROCS(0) doesn't change
	// the current setting but returns the current value
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0), "BUILD", build)

	// -------------------------------------------------
	// Go program receives either a SIGINT or SIGTERM signal,
	// that signal will be sent to the shutdown channel.
	// This allows the program to respond to these signals for graceful shutdown,
	// typically involving cleaning up resources, saving state, or other shutdown procedures.
	// syscall.SIGINT is typically sent when a user types Ctrl+C in the terminal
	// syscall.SIGTERM is a signal often used during system shutdowns or service restarts
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown
	log.Infow("shutdown", "status", "shutdown started", "signal", sig)
	defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

	return nil
}
