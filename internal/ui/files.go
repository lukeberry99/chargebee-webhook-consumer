package ui

import (
	"fmt"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
	"strings"
)

func (ui *UI) loadInitialFiles() {
	initialFiles, err := ui.store.ListEvents()
	if err != nil {
		panic(err) //TODO: Let's not panic
	}

	for _, file := range initialFiles {
		ui.addFileToList(file)
	}
}

func (ui *UI) watchFileUpdates() {
	go func() {
		updates := ui.store.WatchEvents()
		for files := range updates {
			ui.app.QueueUpdateDraw(func() {
				ui.requestList.Clear()
				for _, file := range files {
					ui.addFileToList(file)
				}
			})
		}
	}()
}

func (ui *UI) addFileToList(file storage.EventListItem) {
	ui.requestList.AddItem(file.Filename, file.ReceivedAt, 0, func() {
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
				ui.logView.SetText(ui.logView.GetText(true) + logMsg + "\n")
			})
		}
	}()
}
