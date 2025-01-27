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
	// Generate the appropriate templates for the entire server
	dltemplates, err := s.generateTemplates()
	if err != nil {
		return err
	}
	s.templates = dltemplates

	// Configure a cache buster if the option is enabled
	if !s.DisableCacheBuster {
		s.cacheBuster = s.version
	}

	// Set up an initial server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.router(),
	}

	// Create a signal to wait for an error
	done := make(chan error, 1)

	// Start the server asynchronously
	go func() {
		fmt.Fprintln(s.LogOutput, "Starting server...")
		if err := srv.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				done <- fmt.Errorf("unable to start server: %w", err)
			} else {
				fmt.Fprintln(s.LogOutput, "Server closed. Bye!")
			}
		}
	}()

	// Wait for a closing signal
	go func() {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		<-ctx.Done()
		fmt.Fprintln(s.LogOutput, "Requesting server to stop. Please wait...")

		nctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		done <- srv.Shutdown(nctx)
	}()

	return <-done
}
