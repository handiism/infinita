package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/handiism/infinita/internal/application/port/input"
)

type Server struct {
	httpServer *http.Server
	port       int
	errCh      chan error
}

func New(
	transactions input.TransactionUseCase,
	categories input.CategoryUseCase,
	budgets input.BudgetUseCase,
	reports input.ReportUseCase,
	settings input.SettingsUseCase,
) *Server {
	handler := NewHandler(transactions, categories, budgets, reports, settings)
	mux := NewRouter(handler)

	return &Server{
		errCh: make(chan error, 1),
		httpServer: &http.Server{
			Handler: mux,
		},
	}
}

func (s *Server) Errors() <-chan error {
	return s.errCh
}

func (s *Server) Start(ctx context.Context) (<-chan int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}
	return s.StartOnListener(ctx, listener)
}

func (s *Server) StartOnListener(ctx context.Context, listener net.Listener) (<-chan int, error) {
	if listener == nil {
		return nil, fmt.Errorf("listener is nil")
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("listener address must be TCP, got %T", listener.Addr())
	}

	s.port = addr.Port
	s.errCh = make(chan error, 1)

	portCh := make(chan int, 1)

	// Serve requests and report fatal errors to callers via s.errCh.
	go func() {
		portCh <- s.port
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
			select {
			case s.errCh <- err:
			default:
			}
		}
	}()

	// Best-effort shutdown when the context is cancelled.
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpServer.Shutdown(shutdownCtx)
	}()

	return portCh, nil
}

func (s *Server) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

func (s *Server) Port() int {
	return s.port
}

func (s *Server) WaitForReady(ctx context.Context, baseURL string) error {
	client := &http.Client{Timeout: 100 * time.Millisecond}

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	maxAttempts := 100
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server readiness check cancelled: %w", ctx.Err())
		case <-ticker.C:
			attempts++
			if attempts > maxAttempts {
				return fmt.Errorf("server readiness check timed out after %d attempts", maxAttempts)
			}

			url := fmt.Sprintf("%s/health", baseURL)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				continue
			}
			if err := resp.Body.Close(); err != nil {
				log.Printf("server readiness check: failed to close health response body: %v", err)
			}

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

func (s *Server) Err() <-chan error {
	return s.errCh
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
