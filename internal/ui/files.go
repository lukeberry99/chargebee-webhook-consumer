package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/lukeberry99/webhook-consumer/internal/storage"
)

func (ui *UI) loadInitialFiles() {
	ui.refreshFileList()
}

func (ui *UI) watchFileUpdates() {
	go func() {
		updates := ui.store.WatchEvents()
		for range updates {
			ui.app.QueueUpdateDraw(func() {
				ui.refreshFileList()
			})
		}
	}()
}

func (ui *UI) addFileToList(file storage.EventListItem) {
	secondaryText := file.ReceivedAt
	if file.ServiceName != "" && ui.selectedService == "All" {
		secondaryText = fmt.Sprintf("%s | Service: %s", file.ReceivedAt, file.ServiceName)
	}

	ui.requestList.AddItem(file.Filename, secondaryText, 0, func() {
		content, err := ui.store.ReadEvent(file.Filename)
		if err != nil {
			ui.requestDetails.SetText(fmt.Sprintf("Error reading file: %v", err))
			return
		}

		// Add syntax highlighting for JSON keys
		coloredContent := colorJSONKeys(string(content))
		ui.requestDetails.SetText(coloredContent)
	})
}

// This is a poor implementation, it would be better to support
// something like treesitter and actual colorschemes
func colorJSONKeys(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "\"") {
			// Find the position of the colon to identify keys
			colonPos := strings.Index(line, ":")
			if colonPos > 0 {
				// Find the key portion (everything before the colon)
				beforeColon := line[:colonPos]
				afterColon := line[colonPos:]

				// Add color to the entire key including quotes
				if strings.Contains(beforeColon, "\"") {
					beforeColon = "[#00ffff]" + beforeColon + "[-:-:-]"
				}

				lines[i] = beforeColon + afterColon
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (ui *UI) watchLogs(logChan <-chan string) {
	go func() {
		for logMsg := range logChan {
			ui.app.QueueUpdateDraw(func() {
				if ui.selectedService == "All" || strings.Contains(logMsg, ui.selectedService) {
					currentText := ui.logView.GetText(true)
					ui.logView.SetText(currentText + logMsg + "\n")
					ui.logView.ScrollToEnd()
				}
			})
		}
	}()
}

func (ui *UI) refreshFileList() {
	ui.requestList.Clear()
	files, err := ui.store.ListEvents()
	if err != nil {
		log.Fatalf("failed to load files: %v", err)
	}

	for _, file := range files {
		ui.addFileToList(file)
	}
}
