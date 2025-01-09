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
	"github.com/lukeberry99/webhook-consumer/internal/ngrok"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Error when loading configuration file: %v", err)
	}

	url, err := ngrok.Start()
	if err != nil {
		log.Println("Unable to start ngrok, are you sure it's installed and authenticated?")
		log.Fatalf("Error starting Ngrok: %v", err)
	}

	store, err := storage.NewFileStorage("./logs")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	fmt.Printf("Ngrok URL: %s\n", url)

	port := strconv.Itoa(cfg.Server.Port)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.WebhookHandler(w, r, store)
		}),
	}

	log.Printf("Starting server on port %s", port)

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
	}
}
