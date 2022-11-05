package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (s *Server) ListenAndServe() error {
	// Create a OS Signal handler
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Create a channel to hold closure
	close := make(chan error, 1)

	// Set up an initial server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.handler(),
	}

	// Start the server asynchronously
	go func() {
		fmt.Fprintln(s.LogOutput, "Starting server...")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				close <- err
			} else {
				fmt.Fprintln(s.LogOutput, "Server closed. Bye!")
			}
		}
	}()

	// Wait for a closing signal
	go func() {
		<-sigs

		fmt.Fprintln(s.LogOutput, "Requesting server to stop. Please wait...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		close <- srv.Shutdown(ctx)
	}()

	// Hold here until close happens
	return <-close
}
