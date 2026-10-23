package chainsuite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestCheckEndpointRespectsContext verifies that CheckEndpoint honors the
// caller's context. A stuck RPC endpoint (the server never writes a response)
// must not cause CheckEndpoint to hang: once the context is cancelled the
// underlying HTTP request is aborted and CheckEndpoint returns promptly.
func TestCheckEndpointRespectsContext(t *testing.T) {
	// blocking is closed only after the test ends, so the handler never
	// produces a response. This simulates an RPC endpoint that hangs.
	blocking := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-blocking
	}))
	defer srv.Close()
	defer close(blocking)

	ctx, cancel := context.WithCancel(context.Background())
	// Simulate the suite's outer deadline / cancel firing.
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := CheckEndpoint(ctx, srv.URL, func([]byte) error { return nil })
	if err == nil {
		t.Fatal("expected an error after context cancellation, got nil")
	}
	// Allow a generous slack; the critical behavior is that CheckEndpoint does
	// not block until the test itself times out (default 10m in CI).
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("CheckEndpoint did not respect context cancellation in time: took %s", elapsed)
	}
}

// TestCheckEndpointSuccess verifies the normal path still works: a request
// that returns a body is passed to f and its result is returned.
func TestCheckEndpointSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	called := false
	err := CheckEndpoint(context.Background(), srv.URL, func(b []byte) error {
		called = true
		if string(b) != "hello" {
			t.Fatalf("unexpected body: %q", string(b))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("callback f was not invoked")
	}
}
