package ui

import (
	"os"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (ui *UI) setupKeyBindings() {
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			ui.app.Stop()
		case tcell.KeyTab:
			return ui.handleTabKey()
		}

		if event.Rune() == 'e' {
			return ui.openInEditor()
		}

		if ui.app.GetFocus() == ui.requestList {
			return ui.handleListNavigation(event)
		}

		return event
	})
}

func (ui *UI) openInEditor() *tcell.EventKey {
	currentIndex := ui.requestList.GetCurrentItem()
	if currentIndex < 0 {
		return nil
	}

	filename, _ := ui.requestList.GetItemText(currentIndex)
	if filename == "" {
		return nil
	}

	fullPath := ui.store.GetFullPath(filename)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// The logic below should probably live elsewhere
	// Stop the UI temporarily but keep the webhook server running
	ui.app.Stop()

	// Run the editor
	cmd := exec.Command(editor, fullPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	ui.app = tview.NewApplication()

	ui.initComponents()
	ui.setupLayout()
	ui.setupKeyBindings()
	ui.loadInitialFiles()

	// Restore the previous selection
	ui.requestList.SetCurrentItem(currentIndex)

	// Restart the UI
	if err := ui.app.Run(); err != nil {
		panic(err) //TODO: Let's not panic
	}

	return nil
}

func (ui *UI) handleTabKey() *tcell.EventKey {
	if ui.app.GetFocus() == ui.requestList {
		ui.app.SetFocus(ui.requestDetails)
		ui.statusBar.SetText(" ESC: Quit | j/k/↑/↓: Navigate | TAB: Switch Panel | e: Edit")
	} else {
		ui.app.SetFocus(ui.requestList)
		ui.statusBar.SetText(" ESC: Quit | j/k/↑/↓: Navigate | TAB: Switch Panel | ENTER: View Log | e: Edit")
	}

	return nil
}

func (ui *UI) handleListNavigation(event *tcell.EventKey) *tcell.EventKey {
	switch {
	case event.Rune() == 'j' || event.Key() == tcell.KeyDown:
		current := ui.requestList.GetCurrentItem()
		if current < ui.requestList.GetItemCount()-1 {
			ui.requestList.SetCurrentItem(current + 1)
		}
		return nil
	case event.Rune() == 'k' || event.Key() == tcell.KeyUp:
		current := ui.requestList.GetCurrentItem()
		if current > 0 {
			ui.requestList.SetCurrentItem(current - 1)
		}
		return nil
	case event.Rune() == 'e':
		return ui.openInEditor()
	}
	return event
}
