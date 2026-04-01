package server

import (
	"context"
	"fmt"
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
		httpServer: &http.Server{
			Handler: mux,
		},
	}
}

func (s *Server) Start(ctx context.Context) (<-chan int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	s.port = listener.Addr().(*net.TCPAddr).Port

	portCh := make(chan int, 1)
	s.errCh = make(chan error, 1)

	go func() {
		portCh <- s.port
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.errCh <- err
		}
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
			_ = resp.Body.Close()

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
