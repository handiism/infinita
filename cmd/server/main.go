package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/handiism/infinita/internal/bootstrap"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	dataDir, err := bootstrap.ResolveDataDir()
	if err != nil {
		log.Fatalf("determine data directory: %v", err)
	}

	settingsFile := bootstrap.ResolveSettingsFile(dataDir)

	runtime, err := bootstrap.NewRuntime(context.Background(), dataDir, settingsFile)
	if err != nil {
		log.Fatalf("runtime: %v", err)
	}
	defer func() {
		if closeErr := runtime.Close(); closeErr != nil {
			log.Printf("database close error: %v", closeErr)
		}
	}()

	ln, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
	}
	log.Printf("Starting server on http://%s", ln.Addr().String())

	if _, err := runtime.Server.StartOnListener(ctx, ln); err != nil {
		log.Fatalf("server start error: %v", err)
	}

	select {
	case <-ctx.Done():
	case err := <-runtime.Server.Errors():
		log.Fatalf("server error: %v", err)
	}
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := runtime.Server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}
