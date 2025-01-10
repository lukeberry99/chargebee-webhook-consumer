package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (ui *UI) initComponents() {
	ui.requestList = tview.NewList()
	ui.requestList.
		SetTitle("Requests").
		SetBorder(true)

	ui.requestDetails = tview.NewTextArea()
	ui.requestDetails.
		SetText("Select a request to view details", true).
		SetTitle("Request Details").
		SetBorder(true)

	ui.logView = tview.NewTextArea()
	ui.logView.
		SetTitle("Output").
		SetBorder(true)

	ui.statusBar = tview.NewTextView().
		SetText(" ESC: Quit | j/k: Navigate | TAB: Switch Panel").
		SetTextColor(tcell.ColorYellow)
}

func (ui *UI) setupLayout() {
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(ui.requestList, 0, 1, true).
			AddItem(ui.requestDetails, 0, 2, false),
			0, 2, true).
		AddItem(ui.logView, 0, 1, false).
		AddItem(ui.statusBar, 1, 0, false)

	ui.app.SetRoot(flex, true).EnableMouse(true)
}
