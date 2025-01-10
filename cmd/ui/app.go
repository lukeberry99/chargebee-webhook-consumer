package ui

import (
	"github.com/lukeberry99/webhook-consumer/internal/config"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
	"github.com/rivo/tview"
)

type UI struct {
	app            *tview.Application
	requestList    *tview.List
	requestDetails *tview.TextArea
	logView        *tview.TextArea
	statusBar      *tview.TextView
	store          *storage.FileStorage
	config         *config.Config
}

func New(cfg *config.Config, store *storage.FileStorage) *UI {
	return &UI{
		app:    tview.NewApplication(),
		store:  store,
		config: cfg,
	}
}

func StartUI(cfg *config.Config, logChan <-chan string, store *storage.FileStorage) {
	ui := New(cfg, store)
	ui.initComponents()
	ui.setupLayout()
	ui.setupKeyBindings()
	ui.loadInitialFiles()
	ui.watchFileUpdates()
	ui.watchLogs(logChan)

	if err := ui.app.Run(); err != nil {
		panic(err) //TODO: Maybe let's not panic here
	}
}
