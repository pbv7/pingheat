package pprof

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"
)

// Server provides pprof endpoints.
type Server struct {
	addr   string
	server *http.Server
}

// NewServer creates a new pprof server.
func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

// Start starts the pprof HTTP server.
func (s *Server) Start(ctx context.Context) error {
	s.server = s.newServer()

	go func() {
		<-ctx.Done()
		s.server.Shutdown(context.Background())
	}()

	err := s.server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *Server) newServer() *http.Server {
	mux := http.NewServeMux()

	// Register pprof handlers
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
