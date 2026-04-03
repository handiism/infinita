package server

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

type stubListener struct{}

func (stubListener) Accept() (net.Conn, error) { return nil, fmt.Errorf("not implemented") }
func (stubListener) Close() error              { return nil }
func (stubListener) Addr() net.Addr            { return stubAddr("stub") }

type stubAddr string

func (s stubAddr) Network() string { return string(s) }
func (s stubAddr) String() string  { return string(s) }

func TestStartOnListenerUsesProvidedListener(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := New(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	portCh, err := server.StartOnListener(ctx, listener)
	if err != nil {
		t.Fatalf("StartOnListener() error = %v", err)
	}

	var port int
	select {
	case port = <-portCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server port")
	}

	wantPort := listener.Addr().(*net.TCPAddr).Port
	if port != wantPort {
		t.Fatalf("StartOnListener() port = %d, want %d", port, wantPort)
	}
	if server.Port() != wantPort {
		t.Fatalf("server.Port() = %d, want %d", server.Port(), wantPort)
	}
	if server.Addr() != fmt.Sprintf("127.0.0.1:%d", wantPort) {
		t.Fatalf("server.Addr() = %q, want %q", server.Addr(), fmt.Sprintf("127.0.0.1:%d", wantPort))
	}

	readyCtx, readyCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer readyCancel()
	if err := server.WaitForReady(readyCtx, "http://"+server.Addr()); err != nil {
		t.Fatalf("WaitForReady() error = %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}

func TestStartOnListenerRejectsNilListener(t *testing.T) {
	server := New(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	if _, err := server.StartOnListener(context.Background(), nil); err == nil {
		t.Fatal("StartOnListener() error = nil, want non-nil")
	}
}

func TestStartOnListenerRejectsNonTCPListener(t *testing.T) {
	server := New(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	if _, err := server.StartOnListener(context.Background(), stubListener{}); err == nil {
		t.Fatal("StartOnListener() error = nil, want non-nil")
	}
}

func TestStartOnListenerReportsServeErrors(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	server := New(
		stubTransactionUseCase{},
		stubCategoryUseCase{},
		stubBudgetUseCase{},
		stubReportUseCase{},
		stubSettingsUseCase{},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	portCh, err := server.StartOnListener(ctx, listener)
	if err != nil {
		t.Fatalf("StartOnListener() error = %v", err)
	}

	select {
	case <-portCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for server port")
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close() error = %v", err)
	}

	select {
	case err := <-server.Err():
		if err == nil {
			t.Fatal("Server.Err() returned nil error")
		}
		if !strings.Contains(err.Error(), "closed") {
			t.Fatalf("Server.Err() error = %v, want closed-listener error", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for serve error")
	}
}
