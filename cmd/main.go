package main

import (
	"log"

	"github.com/lukeberry99/webhook-consumer/cmd/ui"
	webhookserver "github.com/lukeberry99/webhook-consumer/cmd/webhook-server"
	"github.com/lukeberry99/webhook-consumer/internal/config"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Error when loading configuration file: %v", err)
	}

	store, err := storage.NewFileStorage("./logs")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
		return
	}

	logChan := make(chan string)

	go webhookserver.StartWebhookServer(cfg, logChan, store)

	ui.StartUI(cfg, logChan, store)
}
