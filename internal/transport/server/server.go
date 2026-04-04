package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/handiism/infinita/internal/application/port/input"
)

// Server is an embedded HTTP server.
type Server struct {
	httpServer *http.Server
	port       int
	errCh      chan error
	startOnce  sync.Once
	startErr   error
}

// New creates a new Server with the given use cases and optional API key.
func New(
	transactions input.TransactionUseCase,
	categories input.CategoryUseCase,
	budgets input.BudgetUseCase,
	reports input.ReportUseCase,
	settings input.SettingsUseCase,
	apiKey string,
) *Server {
	handler := NewHandler(transactions, categories, budgets, reports, settings)

	// Create two muxes: one for health (no auth), one for everything else (with auth)
	healthMux := NewHealthRouter(handler)
	apiMux := NewAPIRouter(handler)

	// Create top-level handler that routes /health to public mux, everything else to auth-wrapped mux
	var topHandler http.Handler
	if apiKey != "" {
		topHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				healthMux.ServeHTTP(w, r)
			} else {
				APIKeyMiddleware(apiKey, apiMux).ServeHTTP(w, r)
			}
		})
	} else {
		topHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				healthMux.ServeHTTP(w, r)
			} else {
				apiMux.ServeHTTP(w, r)
			}
		})
	}

	return &Server{
		errCh: make(chan error, 1),
		httpServer: &http.Server{
			Handler: topHandler,
		},
	}
}

func (s *Server) Start(ctx context.Context) (<-chan int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}
	portCh, err := s.StartOnListener(ctx, listener)
	if err != nil {
		_ = listener.Close()
		return nil, err
	}
	return portCh, nil
}

func (s *Server) StartOnListener(ctx context.Context, listener net.Listener) (<-chan int, error) {
	if listener == nil {
		return nil, fmt.Errorf("listener is nil")
	}

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, fmt.Errorf("listener address must be TCP, got %T", listener.Addr())
	}

	portCh := make(chan int, 1)

	s.startOnce.Do(func() {
		s.port = addr.Port

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
	})

	if s.startErr != nil {
		return nil, s.startErr
	}

	return portCh, nil
}

func (s *Server) Addr() string {
	return fmt.Sprintf("127.0.0.1:%d", s.port)
}

func (s *Server) Port() int {
	return s.port
}

const (
	healthCheckTimeout     = 100 * time.Millisecond
	healthCheckInterval    = 50 * time.Millisecond
	healthCheckMaxAttempts = 100
)

func (s *Server) WaitForReady(ctx context.Context, baseURL string) error {
	client := &http.Client{Timeout: healthCheckTimeout}

	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("server readiness check cancelled: %w", ctx.Err())
		case <-ticker.C:
			attempts++
			if attempts > healthCheckMaxAttempts {
				return fmt.Errorf("server readiness check timed out after %d attempts", healthCheckMaxAttempts)
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
				continue
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
	var shutdownCtx context.Context
	var cancel context.CancelFunc

	if deadline, ok := ctx.Deadline(); ok {
		shutdownCtx, cancel = context.WithDeadline(context.Background(), deadline)
	} else {
		shutdownCtx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	}
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
