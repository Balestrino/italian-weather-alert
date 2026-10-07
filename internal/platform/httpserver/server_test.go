package httpserver

import (
	"context"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// Exercise the real HTTP deadlines without opening a network listener.
func TestServeReturnsResponseAfterTenSecondQuery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()
	listener := &pipeListener{connections: make(chan net.Conn, 1), closed: make(chan struct{})}
	listener.connections <- serverConn
	done := make(chan error, 1)
	go func() {
		done <- Serve(ctx, listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timer := time.NewTimer(11 * time.Second)
			defer timer.Stop()
			select {
			case <-timer.C:
				Reply(w, http.StatusOK, map[string]string{"status": "complete"})
			case <-r.Context().Done():
			}
		}))
	}()
	transport := &http.Transport{DialContext: func(context.Context, string, string) (net.Conn, error) { return clientConn, nil }}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 20 * time.Second}
	request, err := http.NewRequest(http.MethodGet, "http://in-memory.test/situation", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Close = true
	response, err := client.Do(request)
	if err != nil {
		t.Fatalf("complete query response was dropped: %v", err)
	}
	body, err := io.ReadAll(response.Body)
	response.Body.Close()
	if err != nil || response.StatusCode != http.StatusOK || !strings.Contains(string(body), `"complete"`) {
		t.Fatalf("query response: %s %v", body, err)
	}
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

type pipeListener struct {
	connections chan net.Conn
	closed      chan struct{}
	once        sync.Once
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case connection := <-l.connections:
		return connection, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}
func (l *pipeListener) Close() error   { l.once.Do(func() { close(l.closed) }); return nil }
func (l *pipeListener) Addr() net.Addr { return pipeAddress{} }

type pipeAddress struct{}

func (pipeAddress) Network() string { return "pipe" }
func (pipeAddress) String() string  { return "in-memory" }
