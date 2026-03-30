package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/patrickdappollonio/http-server/internal/middlewares"
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

	if s.IsTLSEnabled() {
		return s.listenTLS()
	}

	return s.listenHTTPOnly()
}

// listenHTTPOnly starts a single HTTP listener (existing behavior).
func (s *Server) listenHTTPOnly() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.router(),
	}

	done := make(chan error, 1)

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

// listenTLS starts dual HTTP+HTTPS listeners with errgroup coordination.
func (s *Server) listenTLS() error {
	router := s.router()

	// HTTPS server with GetCertificate callback
	tlsConfig := &tls.Config{
		GetCertificate: s.getCertificate,
	}
	httpsServer := &http.Server{
		Addr:      fmt.Sprintf(":%d", s.HTTPSPort),
		Handler:   router,
		TLSConfig: tlsConfig,
	}

	// HTTP redirect server (disabled when --http-port 0)
	var httpServer *http.Server
	if s.HTTPPort != 0 {
		httpServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", s.HTTPPort),
			Handler: s.httpRedirectHandler(),
		}
	}

	g, ctx := errgroup.WithContext(context.Background())

	// HTTPS listener
	g.Go(func() error {
		fmt.Fprintf(s.LogOutput, "Starting HTTPS server on :%d...\n", s.HTTPSPort)
		// Empty cert/key paths: GetCertificate provides the cert
		if err := httpsServer.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			return fmt.Errorf("HTTPS server error: %w", err)
		}
		return nil
	})

	// HTTP redirect listener
	if httpServer != nil {
		g.Go(func() error {
			fmt.Fprintf(s.LogOutput, "Starting HTTP server on :%d (redirecting to HTTPS)...\n", s.HTTPPort)
			if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
				return fmt.Errorf("HTTP server error: %w", err)
			}
			return nil
		})
	}

	// Context watcher: if the errgroup context is cancelled (from any goroutine
	// failing), shut down all servers to prevent g.Wait() from hanging.
	g.Go(func() error {
		<-ctx.Done()
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutCancel()
		httpsServer.Shutdown(shutCtx) //nolint:errcheck,contextcheck // best-effort shutdown with fresh deadline
		if httpServer != nil {
			httpServer.Shutdown(shutCtx) //nolint:errcheck,contextcheck // best-effort shutdown with fresh deadline
		}
		return nil
	})

	// Signal handler: SIGINT/SIGTERM trigger graceful shutdown
	g.Go(func() error {
		sigCtx, sigCancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		defer sigCancel()

		<-sigCtx.Done()
		fmt.Fprintln(s.LogOutput, "Requesting server to stop. Please wait...")

		//nolint:contextcheck // intentionally creating a new context for shutdown deadline, independent of the cancelled signal context
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer shutCancel()

		if err := httpsServer.Shutdown(shutCtx); err != nil { //nolint:contextcheck // shutdown uses a fresh deadline context, not the cancelled signal context
			fmt.Fprintf(s.LogOutput, "HTTPS shutdown error: %v\n", err)
		}
		if httpServer != nil {
			if err := httpServer.Shutdown(shutCtx); err != nil { //nolint:contextcheck // same as above
				fmt.Fprintf(s.LogOutput, "HTTP shutdown error: %v\n", err)
			}
		}

		fmt.Fprintln(s.LogOutput, "Server closed. Bye!")
		return nil
	})

	// SIGHUP handler for cert reload (Unix only, no-op on Windows)
	s.startSIGHUPHandler(ctx)

	//nolint:wrapcheck // errors from errgroup already wrapped by listener goroutines
	return g.Wait()
}

// httpRedirectHandler returns an HTTP handler that redirects all requests
// to the HTTPS equivalent using the configured hostname.
func (s *Server) httpRedirectHandler() http.Handler {
	return middlewares.LogRequest(s.LogOutput, logFormat, "token")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + s.Hostname
			if s.HTTPSPort != 443 {
				target += fmt.Sprintf(":%d", s.HTTPSPort)
			}
			target += r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		}),
	)
}
