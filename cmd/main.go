package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lukeberry99/webhook-consumer/internal/config"
	"github.com/lukeberry99/webhook-consumer/internal/server"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
	"github.com/lukeberry99/webhook-consumer/internal/tunnel"
	"github.com/lukeberry99/webhook-consumer/internal/ui"
)

func main() {
	logChan := make(chan string, 100) // Buffered channel to prevent blocking
	logChan <- "Starting webhook consumer..."

	cfg, err := config.Load("")
	if err != nil {
		logChan <- fmt.Sprintf("Error when loading configuration file: %v", err)
	}

	storageDir := cfg.Storage.Path
	store, err := storage.NewFileStorage(storageDir)
	if err != nil {
		logChan <- fmt.Sprintf("Failed to create storage: %v", err)
	}

	logChan <- "Initialising UI..."
	uiDone := make(chan struct{})
	uiErr := make(chan error, 1)
	go func() {
		if err := ui.StartUI(cfg, logChan, store); err != nil {
			logChan <- fmt.Sprintf("UI Error: %v", err)
			uiErr <- err
			close(uiDone)
			return
		}

		close(uiDone)
	}()

	// Start monitoring UI errors in background
	go func() {
		if err := <-uiErr; err != nil {
			fmt.Printf("Failed to start UI: %v\n", err)
			os.Exit(1)
		}
	}()

	// Check if terminal is available
	if _, err := os.Stdout.Stat(); err != nil {
		fmt.Printf("Terminal not available: %v\n", err)
		os.Exit(1)
	}

	// Give the UI some time to initialize
	time.Sleep(100 * time.Millisecond)
	logChan <- "UI initialised successfully"

	var tunnelServer tunnel.Tunnel
	if cfg.Tunnel.Driver != "local" {
		tunnelConfig := tunnel.Config{
			Provider:        tunnel.Provider(cfg.Tunnel.Driver),
			CloudflareToken: cfg.Tunnel.CloudflareToken,
		}

		var tunnelErr error
		tunnelServer, tunnelErr = tunnel.New(tunnelConfig)
		if tunnelErr != nil {
			logChan <- fmt.Sprintf("Failed to create tunnel: %v - falling back to local mode", tunnelErr)
			cfg.Tunnel.Driver = "local"
		} else {
			tunnelURL, startErr := tunnelServer.Start()
			if startErr != nil {
				logChan <- fmt.Sprintf("Failed to start tunnel: %v - falling back to local mode", startErr)
				cfg.Tunnel.Driver = "local"
				tunnelServer = nil
			} else {
				logChan <- fmt.Sprintf("Tunnel URL: %s", tunnelURL)
			}
		}
	}

	if cfg.Tunnel.Driver == "local" {
		url := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		logChan <- "Running in local mode - no tunnel started"
		logChan <- fmt.Sprintf("Tunnel URL: %s", url)
	}

	srv := server.NewWebhookServer(cfg, store, logChan)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logChan <- fmt.Sprintf("Error starting server: %v", err)
	case sig := <-shutdown:
		logChan <- fmt.Sprintf("Starting shutdown, received signal: %v", sig)
	case <-uiDone:
		logChan <- "UI closed, shutting down server"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logChan <- fmt.Sprintf("Error during shutdown: %v", err)
		if err := srv.Close(); err != nil {
			logChan <- fmt.Sprintf("Error during forced shutdown: %v", err)
		}
	}

	if tunnelServer != nil {
		if err := tunnelServer.Stop(); err != nil {
			logChan <- fmt.Sprintf("Error stopping tunnel: %v", err)
		}
	}

	close(logChan)
}
