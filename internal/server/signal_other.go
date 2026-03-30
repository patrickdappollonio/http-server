//go:build windows

package server

import "context"

// startSIGHUPHandler is a no-op on Windows. Use the /_/tls/reload
// endpoint for certificate reloading instead.
func (s *Server) startSIGHUPHandler(_ context.Context) {}
