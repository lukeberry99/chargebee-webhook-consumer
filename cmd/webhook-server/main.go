package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lukeberry99/webhook-consumer/internal/config"
	"github.com/lukeberry99/webhook-consumer/internal/handler"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
	"github.com/lukeberry99/webhook-consumer/internal/tunnel"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Error when loading configuration file: %v", err)
	}

	port := strconv.Itoa(cfg.Server.Port)

	var tunnelURL string
	var tunnelServer tunnel.Tunnel

	if cfg.Tunnel.Driver != "local" {
		tunnelConfig := tunnel.Config{
			Provider:        tunnel.Provider(cfg.Tunnel.Driver),
			CloudflareToken: cfg.Tunnel.CloudflareToken,
		}

		tunnelService, err := tunnel.New(tunnelConfig)
		if err != nil {
			log.Fatalf("Failed to create tunnel: %v", err)
		}

		tunnelURL, err = tunnelService.Start()
		if err != nil {
			log.Fatalf("Failed to start tunnel: %v", err)
		}

		log.Printf("Tunnel URL: %s", tunnelURL)
		defer tunnelServer.Stop()
	} else {
		url := fmt.Sprintf("http://localhost:%s", port)
		log.Println("Running in local mode - no tunnel started")
		log.Printf("Tunnel URL: %s\n", url)
	}

	store, err := storage.NewFileStorage("./logs")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	srv := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.WebhookHandler(w, r, store)
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
		log.Fatalf("Error starting server: %v", err)

	case sig := <-shutdown:
		log.Printf("Starting shutdown, received signal: %v", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			if err := srv.Close(); err != nil {
				log.Printf("Error during forced shutdown: %v", err)
			}
		}

		if tunnelServer != nil {
			if err := tunnelServer.Stop(); err != nil {
				log.Printf("Error stopping tunnel: %v", err)
			}
		}
	}
}
