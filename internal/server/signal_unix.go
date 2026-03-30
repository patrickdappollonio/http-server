//go:build !windows

package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// startSIGHUPHandler starts a goroutine that reloads TLS certificates
// on SIGHUP with a 1-second debounce to prevent rapid reload storms.
func (s *Server) startSIGHUPHandler(ctx context.Context) {
	if !s.IsTLSEnabled() {
		return
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP)
		defer signal.Stop(sigCh)

		var lastReload time.Time

		for {
			select {
			case <-ctx.Done():
				return
			case <-sigCh:
				if time.Since(lastReload) < time.Second {
					fmt.Fprintln(s.LogOutput, "SIGHUP received but debounced (less than 1s since last reload)")
					continue
				}
				lastReload = time.Now()
				if err := s.reloadCert(); err != nil {
					fmt.Fprintf(s.LogOutput, "TLS certificate reload failed: %v\n", err)
				} else {
					fmt.Fprintln(s.LogOutput, "TLS certificate reloaded successfully")
				}
			}
		}
	}()
}
