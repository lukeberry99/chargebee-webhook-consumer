package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (ui *UI) initComponents() {
	ui.requestList = tview.NewList()
	ui.requestList.
		SetMainTextStyle(tcell.StyleDefault).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue)).
		SetHighlightFullLine(true).
		SetSecondaryTextColor(tcell.ColorGray).
		SetTitle("Requests").
		SetBorder(true)

	ui.requestDetails = tview.NewTextView()
	ui.requestDetails.
		SetDynamicColors(true).
		SetWrap(true).
		SetText("Select a request to view details").
		SetTitle("Request Details").
		SetBorder(true)

	ui.logView = tview.NewTextView()
	ui.logView.
		SetTitle("Output").
		SetBorder(true)

	ui.statusBar = tview.NewTextView().
		SetText(" ESC: Quit | j/k/↑/↓: Navigate | TAB: Switch Panel | ENTER: View Log | e: Edit").
		SetTextColor(tcell.ColorYellow)
}

func (ui *UI) setupLayout() {
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(ui.requestList, 0, 1, true).
			AddItem(ui.requestDetails, 0, 4, false), // 20%:80% split
			0, 2, true).
		AddItem(ui.logView, 0, 1, false).
		AddItem(ui.statusBar, 1, 0, false)

	ui.app.SetRoot(flex, true).EnableMouse(true)
}
