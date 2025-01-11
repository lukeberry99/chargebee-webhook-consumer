package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (ui *UI) initComponents() {
	ui.serviceModal = tview.NewModal().
		SetText("Select Service").
		AddButtons([]string{"All"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "All" {
				ui.store.SetSelectedService("")
				ui.requestList.SetTitle("Requests [yellow](All)[-]")
			} else {
				ui.store.SetSelectedService(buttonLabel)
				ui.requestList.SetTitle(fmt.Sprintf("Requests [yellow](%s)[-]", buttonLabel))
			}
			ui.selectedService = buttonLabel
			ui.refreshFileList()

			ui.app.SetRoot(ui.mainFlex, true)
		})

	var buttons []string
	buttons = append(buttons, "All")
	if ui.config != nil {
		for service := range ui.config.Services {
			buttons = append(buttons, service)
		}
	}
	ui.serviceModal.ClearButtons()
	ui.serviceModal.AddButtons(buttons)

	ui.selectedService = "All"
	// ui.store.SetSelectedService("")

	ui.requestList = tview.NewList()
	ui.requestList.
		SetMainTextStyle(tcell.StyleDefault).
		SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorBlue)).
		SetHighlightFullLine(true).
		SetSecondaryTextColor(tcell.ColorGray).
		SetTitle(fmt.Sprintf("Requests [yellow](%s)[-]", ui.selectedService)).
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
		SetText(" ESC: Quit | j/k/↑/↓: Navigate | TAB: Switch Panel | ENTER: View Log | e: Edit | s: Select Service").
		SetTextColor(tcell.ColorYellow)
}

func (ui *UI) setupLayout() {
	ui.mainFlex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(ui.requestList, 0, 1, true).
			AddItem(ui.requestDetails, 0, 4, false), // 20%:80% split
			0, 2, true).
		AddItem(ui.logView, 0, 1, false).
		AddItem(ui.statusBar, 1, 0, false)

	ui.app.SetRoot(ui.mainFlex, true).EnableMouse(true)
}
