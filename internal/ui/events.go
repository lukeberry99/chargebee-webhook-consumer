package ui

import "github.com/gdamore/tcell/v2"

func (ui *UI) setupKeyBindings() {
	ui.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			ui.app.Stop()
		case tcell.KeyTab:
			return ui.handleTabKey()
		}

		if ui.app.GetFocus() == ui.requestList {
			return ui.handleListNavigation(event)
		}

		return event
	})
}

func (ui *UI) handleTabKey() *tcell.EventKey {
	if ui.app.GetFocus() == ui.requestList {
		ui.app.SetFocus(ui.requestDetails)
		ui.statusBar.SetText(" ESC: Quit | TAB: Switch Panel")
	} else {
		ui.app.SetFocus(ui.requestList)
		ui.statusBar.SetText(" ESC: Quit | j/k: Navigate | TAB: Switch Panel")
	}

	return nil
}

func (ui *UI) handleListNavigation(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'j':
		current := ui.requestList.GetCurrentItem()
		if current < ui.requestList.GetItemCount()-1 {
			ui.requestList.SetCurrentItem(current + 1)
		}
		return nil
	case 'k':
		current := ui.requestList.GetCurrentItem()
		if current > 0 {
			ui.requestList.SetCurrentItem(current - 1)
		}
		return nil
	}
	return event
}
