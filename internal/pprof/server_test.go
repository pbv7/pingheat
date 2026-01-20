package pprof

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestServerHandlersAndTimeouts(t *testing.T) {
	s := NewServer("127.0.0.1:6060")
	server := s.newServer()

	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout=%v, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("ReadTimeout=%v, want 10s", server.ReadTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout=%v, want 60s", server.IdleTimeout)
	}

	req := httptest.NewRequest(http.MethodGet, "/debug/pprof/", nil)
	handler, pattern := server.Handler.(*http.ServeMux).Handler(req)
	if pattern == "" || handler == nil {
		t.Fatalf("expected pprof handler for /debug/pprof/")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("pprof status=%d, want 200", rec.Code)
	}
}
