package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	AlgoModeEnvKey = "MODE"

	AlgoModeMeter = "METER"
	AlgoModeQueue = "QUEUE"
)

func main() {
	// get the right server for the mode
	mode := getMode()
	rootHandler := getHandler(mode)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", "27009"),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  3 * time.Second,
		Handler:      rootHandler,
	}

	errChan := make(chan error)
	go func() {
		log.Printf("Starting server on %q", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	for {
		select {
		case <-sigint:
			log.Println("Shutting down...")
			srv.Shutdown(shutdownCtx)
			return
		case err := <-errChan:
			log.Printf("An error occurred: %v\n", err)
			return
		}
	}
}

//getHandler returns the right handler for the mode
func getHandler(mode string) http.Handler {
	if strings.ToUpper(mode) == AlgoModeMeter {
		return handleAsAMeter()
	}
	panic("Queue Not Implemented")
}

//getMode reads the algorithm type to use from the environment.
//Could be "METER" or "QUEUE"
func getMode() string {
	if mode, ok := os.LookupEnv(strings.ToLower(AlgoModeEnvKey)); ok {
		return mode
	}
	panic(fmt.Sprintf("%q not found in environment", AlgoModeEnvKey))
}
