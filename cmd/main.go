package main

import (
	"context"
	"fmt"
	"log"
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
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Error when loading configuration file: %v", err)
	}

	store, err := storage.NewFileStorage("./logs")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	logChan := make(chan string)

	uiDone := make(chan struct{})
	go func() {
		ui.StartUI(cfg, logChan, store)
		close(uiDone)
	}()

	// Give the UI some time to start up
	time.Sleep(100 * time.Millisecond)

	var tunnelServer tunnel.Tunnel
	if cfg.Tunnel.Driver != "local" {
		tunnelConfig := tunnel.Config{
			Provider:        tunnel.Provider(cfg.Tunnel.Driver),
			CloudflareToken: cfg.Tunnel.CloudflareToken,
		}

		tunnelServer, err = tunnel.New(tunnelConfig)
		if err != nil {
			log.Fatalf("Failed to create tunnel: %v", err)
		}

		tunnelURL, err := tunnelServer.Start()
		if err != nil {
			log.Fatalf("Failed to start tunnel: %v", err)
		}

		logChan <- fmt.Sprintf("Tunnel UR: %s", tunnelURL)
	} else {
		url := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		logChan <- fmt.Sprintf("Running in local mode - no tunnel started")
		logChan <- fmt.Sprintf("Tunnel URL: %s", url)
	}

	srv := server.NewWebhookServer(cfg, store, logChan)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	logChan <- fmt.Sprintf("Server started on port %d", cfg.Server.Port)

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
