package ui

import (
	"fmt"

	"github.com/lukeberry99/whook/internal/config"
	"github.com/lukeberry99/whook/internal/storage"
	"github.com/rivo/tview"
)

type UI struct {
	app             *tview.Application
	requestList     *tview.List
	requestDetails  *tview.TextView
	logView         *tview.TextView
	statusBar       *tview.TextView
	serviceModal    *tview.Modal
	mainFlex        *tview.Flex
	store           *storage.FileStorage
	config          *config.Config
	selectedService string
	isModalVisible  bool
}

func New(cfg *config.Config, store *storage.FileStorage) *UI {
	ui := &UI{
		app:    tview.NewApplication(),
		store:  store,
		config: cfg,
	}

	ui.selectedService = "All"
	store.SetSelectedService("")

	return ui
}

func StartUI(cfg *config.Config, logChan <-chan string, store *storage.FileStorage) error {
	ui := New(cfg, store)
	ui.initComponents()
	ui.setupLayout()
	ui.setupKeyBindings()
	ui.loadInitialFiles()
	ui.watchFileUpdates()
	ui.watchLogs(logChan)

	ui.statusBar.SetText(" ESC: Quit | j/k/↑/↓: Navigate Services | ENTER: Select | TAB: Switch Panel | s: Select Service")

	if err := ui.app.Run(); err != nil {
		return fmt.Errorf("failed to start UI: %w", err)
	}

	return nil
}
