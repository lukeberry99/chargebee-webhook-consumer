package ui

import (
	"fmt"
	"github.com/lukeberry99/webhook-consumer/internal/storage"
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
	ui.requestList.AddItem(file.Filename, fmt.Sprintf("Project - %s", file.ReceivedAt), 0, func() {
		content, err := ui.store.ReadEvent(file.Filename)
		if err != nil {
			ui.requestDetails.SetText(fmt.Sprintf("Error reading file: %v", err), true)
			return
		}
		ui.requestDetails.SetText(string(content), true)
	})
}

func (ui *UI) watchLogs(logChan <-chan string) {
	go func() {
		for logMsg := range logChan {
			ui.app.QueueUpdateDraw(func() {
				ui.logView.SetText(ui.logView.GetText()+logMsg+"\n", true)
			})
		}
	}()
}
