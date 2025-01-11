package server

import (
	"fmt"
	"net/http"

	"github.com/lukeberry99/whook/internal/config"
	"github.com/lukeberry99/whook/internal/handler"
	"github.com/lukeberry99/whook/internal/storage"
)

func NewWebhookServer(cfg *config.Config, store *storage.FileStorage, logChan chan<- string) *http.Server {
	return &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handler.WebhookHandler(w, r, store, logChan)
		}),
	}
}
