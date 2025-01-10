package webhookserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lukeberry99/webhook-consumer/internal/config"
	"github.com/lukeberry99/webhook-consumer/internal/handler"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
	"github.com/lukeberry99/webhook-consumer/internal/tunnel"
)

func StartWebhookServer(cfg *config.Config, logChan chan<- string, store *storage.FileStorage) {
	defer close(logChan)

	var tunnelURL string
	var tunnelServer tunnel.Tunnel

	if cfg.Tunnel.Driver != "local" {
		tunnelConfig := tunnel.Config{
			Provider:        tunnel.Provider(cfg.Tunnel.Driver),
			CloudflareToken: cfg.Tunnel.CloudflareToken,
		}

		tunnelService, err := tunnel.New(tunnelConfig)
		if err != nil {
			logChan <- fmt.Sprintf("Failed to create tunnel: %v", err)
			return
		}

		tunnelURL, err = tunnelService.Start()
		if err != nil {
			logChan <- fmt.Sprintf("Failed to start tunnel: %v", err)
			return
		}

		logChan <- fmt.Sprintf("Tunnel URL: %s", tunnelURL)
		defer tunnelServer.Stop()
	} else {
		url := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
		logChan <- fmt.Sprintf("Running in local mode - no tunnel started")
		logChan <- fmt.Sprintf("Tunnel URL: %s\n", url)
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.WebhookHandler(w, r, store, logChan)
		}),
	}

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
	}
}
